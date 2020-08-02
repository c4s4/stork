package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	// RegexpScript is the regexp for scripts to run
	RegexpScript = `^\d+.*\.sql$`
	// QueryMetaExists is the query to check that meta table exists
	QueryMetaExists = `
	SELECT count(*) FROM stork.script;
	`
	// QueryEraseMeta is the query to erase meta table
	QueryEraseMeta = `
	DROP TABLE IF EXISTS stork.script;
	`
	// QueryCreateMeta is the query to create meta table
	QueryCreateMeta = `
	CREATE DATABASE IF NOT EXISTS stork;
	CREATE TABLE stork.script (
		id INTEGER NOT NULL AUTO_INCREMENT,
		name TEXT NOT NULL,
		date DATETIME DEFAULT CURRENT_TIMESTAMP,
		success BOOLEAN NOT NULL,
		error TEXT,
		PRIMARY KEY (id)
	);
	`
	// QueryScriptPassed is the query to determine if a script already passed
	QueryScriptPassed = `
	SELECT count(*)
	  FROM stork.script
	 WHERE name = ?
	   AND success = 1;
	`
	// QueryRecordResult is the query to record script passing
	QueryRecordResult = `
	INSERT INTO stork.script (name, success, error)
	VALUES (?, ?, ?);
	`
)

// Version should be provided at compile time
var Version string
var db *sql.DB
var mute bool
var white bool

// ParseCommandLine does what you think
func ParseCommandLine() (string, string, bool, bool, bool, bool, bool, bool) {
	flag.Usage = func() {
		fmt.Println(`Usage: stork [-env=file] [-init] [-fill] [-dry] [-mute] [-white] [-version] dir
-env=file  Dotenv file to load
-init      Run all scripts
-fill      Fill script table with all migration scripts
-dry       Dry run (won't execute scripts)
-mute      Don't print logs
-white     Don't print color
-version   Print version and exit
dir        Directory of migration scripts`)
	}
	env := flag.String("env", "", "Dotenv file to load")
	init := flag.Bool("init", false, "Run all scripts")
	fill := flag.Bool("fill", false, "Fill script table with all migration scripts")
	mute := flag.Bool("mute", false, "Don't print logs")
	white := flag.Bool("white", false, "Don't print color")
	dry := flag.Bool("dry", false, "Dry run (won't execute scripts)")
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()
	dirs := flag.Args()
	dir := "."
	if len(dirs) > 1 {
		Error("You can pass only one directory")
	}
	if len(dirs) == 1 {
		dir = dirs[0]
	}
	return *env, dir, *init, *fill, *dry, *mute, *white, *version
}

// Print prints given text if not mute
func Print(text string, args ...interface{}) {
	if !mute {
		if args != nil {
			text = fmt.Sprintf(text, args...)
		}
		fmt.Println(text)
	}
}

// Error prints an error message and exits
func Error(text string, args ...interface{}) {
	error := "\033[1;31mERROR\033[0m "
	if white {
		error = "ERROR "
	}
	if args != nil {
		text = fmt.Sprintf(text, args...)
	}
	println(error + text)
	os.Exit(1)
}

// CheckError prints message on error
func CheckError(err error, message string, args ...interface{}) {
	if err != nil {
		args = append(args, err)
		Error(message, args...)
	}
}

// PrintOK prints OK in green
func PrintOK() {
	ok := "\033[1;32mOK\033[0m"
	if white {
		ok = "OK"
	}
	Print(ok)
}

// LoadEnv loads environment in given file
func LoadEnv(filename string) error {
	Print("Loading environment %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		bytes, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		line := strings.TrimSpace(string(bytes))
		if line[0] == '#' {
			continue
		}
		index := strings.Index(line, "=")
		if index < 0 {
			return fmt.Errorf("bad environment line: '%s'", line)
		}
		name := strings.TrimSpace(line[:index])
		value := strings.TrimSpace(line[index+1:])
		os.Setenv(name, value)
	}
	return nil
}

// ConnectDatabase returns database connection
func ConnectDatabase() *sql.DB {
	var err error
	source := os.Getenv("MYSQL_USERNAME") + ":" + os.Getenv("MYSQL_PASSWORD") +
		"@tcp(" + os.Getenv("MYSQL_HOSTNAME") + ")/"
	db, err := sql.Open("mysql", source)
	if err != nil {
		Error("connecting database: %v", err)
	}
	return db
}

// EraseMetaTable initializes meta tables if necessary
func EraseMetaTable() error {
	Print("Erasing meta table")
	return ExecuteScript(QueryEraseMeta)
}

// CreateMetaTable initializes meta tables if necessary
func CreateMetaTable() error {
	err := ExecuteScript(QueryMetaExists)
	if err != nil {
		Print("Creating meta table")
		if err := ExecuteScript(QueryCreateMeta); err != nil {
			return err
		}
	}
	return nil
}

// ScriptsList returns the list of SQL scripts in given directory
func ScriptsList(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var scripts []string
	for _, file := range files {
		match, err := regexp.MatchString(RegexpScript, strings.ToLower(file.Name()))
		if err != nil {
			return nil, err
		}
		if match {
			scripts = append(scripts, file.Name())
		}
	}
	sort.Strings(scripts)
	return scripts, nil
}

// ScriptPassed tells if given script already passed
func ScriptPassed(script string) (bool, error) {
	var count int
	err := db.QueryRow(QueryScriptPassed, script).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExecuteScript executes given script
func ExecuteScript(source string) error {
	for _, query := range strings.Split(string(source), ";\n") {
		query := strings.TrimSpace(query)
		if query != "" {
			_, err := db.Exec(query)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// RunScript runs given script
func RunScript(dir, script string) error {
	Print("Running script %s", script)
	file, err := os.Open(filepath.Join(dir, script))
	if err != nil {
		return err
	}
	defer file.Close()
	source, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, query := range strings.Split(string(source), ";\n") {
		query := strings.TrimSpace(query)
		if query != "" {
			if _, err := db.Exec(query); err != nil {
				tx.Rollback()
				if err = RecordResult(script, err); err != nil {
					return err
				}
				Error("running script %s: %v", script, err)
			}
		}
	}
	tx.Commit()
	RecordResult(script, nil)
	return nil
}

// RecordResult record script result in meta table
func RecordResult(script string, err error) error {
	var success bool
	var message string
	if err != nil {
		success = false
		message = err.Error()
	} else {
		success = true
		message = ""
	}
	if _, err = db.Exec(QueryRecordResult, script, success, message); err != nil {
		return err
	}
	return nil
}

// RunFill fills script table with migrations scripts
func RunFill(dir string) {
	err := EraseMetaTable()
	CheckError(err, "erasing meta table: %v")
	err = CreateMetaTable()
	CheckError(err, "creating meta tables: %v")
	scripts, err := ScriptsList(dir)
	CheckError(err, "getting scripts list: %v")
	for _, script := range scripts {
		Print("Filling script %s", script)
		err = RecordResult(script, err)
		CheckError(err, "recording script: %v")
	}
}

// RunDry runs dry migrations
func RunDry(dir string, init bool) {
	if init {
		Print("Erasing meta table")
		Print("Creating meta table")
	}
	scripts, err := ScriptsList(dir)
	CheckError(err, "getting scripts list: %v")
	for _, script := range scripts {
		passed, err := ScriptPassed(script)
		CheckError(err, "determining if script %s passed: %v", script)
		if !passed || init {
			Print("Running script %s", script)
		} else {
			Print("Skipping script %s", script)
		}
	}
}

// RunMain runs migrations
func RunMain(dir string, init bool) {
	if init {
		err := EraseMetaTable()
		CheckError(err, "erasing meta table: %v")
	}
	scripts, err := ScriptsList(dir)
	CheckError(err, "getting scripts list: %v")
	err = CreateMetaTable()
	CheckError(err, "creating meta tables: %v")
	for _, script := range scripts {
		passed, err := ScriptPassed(script)
		CheckError(err, "determining if script %s passed: %v", script)
		if !passed {
			err = RunScript(dir, script)
			CheckError(err, "running script %s: %v", script)
		} else {
			Print("Skipping script %s", script)
		}
	}
}

func main() {
	var env, dir string
	var init, fill, dry, version bool
	env, dir, init, fill, dry, mute, white, version = ParseCommandLine()
	if version {
		Print(Version)
		os.Exit(0)
	}
	if env != "" {
		err := LoadEnv(env)
		CheckError(err, "loading dotenv file %s: %v", env)
	}
	db = ConnectDatabase()
	defer db.Close()
	if fill {
		RunFill(dir)
	} else if dry {
		RunDry(dir, init)
	} else {
		RunMain(dir, init)
	}
	PrintOK()
}
