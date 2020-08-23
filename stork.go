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
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
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
		date TIMESTAMP NOT NULL,
		PRIMARY KEY (id)
	);
	`
	// QueryScriptPassed is the query to determine if a script already passed
	QueryScriptPassed = `
	SELECT count(*)
	  FROM stork.script
	 WHERE name = ?;
	`
	// QueryRecordResult is the query to record script passing
	QueryRecordResult = `
	INSERT INTO stork.script (name)
	VALUES (?);
	`
)

// Version should be provided at compile time
var Version string
var db *sql.DB
var tx *sql.Tx
var mute bool
var white bool

// RegexpScript is the regexp for scripts to run
var RegexpScript = regexp.MustCompile(`^\d+.*\.sql$`)

// RegexpIndex is the regexp to extract script index
var RegexpIndex = regexp.MustCompile(`^\d+`)

// ParseCommandLine does what you think
func ParseCommandLine() (string, string, string, bool, bool, bool, bool, bool, bool) {
	flag.Usage = func() {
		fmt.Println(`Usage: stork [-env=file] [-init] [-fill] [-dry] [-mute] [-white] [-version] dir
-env=file  Dotenv file to load
-upto=XYZ  Run scripts up to the one starting with XYZ
-init      Run all scripts
-fill      Fill script table with all migration scripts
-dry       Dry run (won't execute scripts)
-mute      Don't print logs
-white     Don't print color
-version   Print version and exit
dir        Directory of migration scripts`)
	}
	env := flag.String("env", "", "Dotenv file to load")
	upto := flag.String("upto", "", "Run scripts up to the one starting with XYZ")
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
	return *env, dir, *upto, *init, *fill, *dry, *mute, *white, *version
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
	tx.Rollback()
	db.Close()
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
func ConnectDatabase() {
	var err error
	source := os.Getenv("MYSQL_USERNAME") + ":" + os.Getenv("MYSQL_PASSWORD") +
		"@tcp(" + os.Getenv("MYSQL_HOSTNAME") + ")/"
	db, err = sql.Open("mysql", source)
	if err != nil {
		Error("connecting database: %v", err)
	}
	tx, err = db.Begin()
	if err != nil {
		Error("starting transaction: %v", err)
	}
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
		if RegexpScript.MatchString(strings.ToLower(file.Name())) {
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
			_, err := tx.Exec(query)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// StrToInt converts string to integer
func StrToInt(str string) int {
	i, err := strconv.Atoi(str)
	CheckError(err, "bad script index '%s': %v", str)
	return i
}

// AfterUpto tells if given script is after upto limit
func AfterUpto(script, upto string) bool {
	if upto == "" {
		return false
	}
	return StrToInt(RegexpIndex.FindString(script)) > StrToInt(upto)
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
	if err := ExecuteScript(string(source)); err != nil {
		Error("running script %s: %v", script, err)
	}
	if _, err := tx.Exec(QueryRecordResult, script); err != nil {
		Error("recording result for script %s: %v", script, err)
	}
	return nil
}

// RunMain runs migrations
func RunMain(dir, upto string, init bool) {
	if init {
		err := EraseMetaTable()
		CheckError(err, "erasing meta table: %v")
	}
	scripts, err := ScriptsList(dir)
	CheckError(err, "getting scripts list: %v")
	err = CreateMetaTable()
	CheckError(err, "creating meta tables: %v")
	for _, script := range scripts {
		if AfterUpto(script, upto) {
			break
		}
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

// RunFill fills script table with migrations scripts
func RunFill(dir, upto string) {
	err := EraseMetaTable()
	CheckError(err, "erasing meta table: %v")
	err = CreateMetaTable()
	CheckError(err, "creating meta tables: %v")
	scripts, err := ScriptsList(dir)
	CheckError(err, "getting scripts list: %v")
	for _, script := range scripts {
		if AfterUpto(script, upto) {
			break
		}
		Print("Filling script %s", script)
		if _, err := tx.Exec(QueryRecordResult, script); err != nil {
			Error("filling script %s: %v", script, err)
		}
	}
}

// RunDry runs dry migrations
func RunDry(dir, upto string, init bool) {
	if init {
		Print("Erasing meta table")
		Print("Creating meta table")
	}
	scripts, err := ScriptsList(dir)
	CheckError(err, "getting scripts list: %v")
	for _, script := range scripts {
		if AfterUpto(script, upto) {
			break
		}
		passed, err := ScriptPassed(script)
		CheckError(err, "determining if script %s passed: %v", script)
		if !passed || init {
			Print("Running script %s", script)
		} else {
			Print("Skipping script %s", script)
		}
	}
}

func main() {
	var env, dir, upto string
	var init, fill, dry, version bool
	env, dir, upto, init, fill, dry, mute, white, version = ParseCommandLine()
	if version {
		Print(Version)
		os.Exit(0)
	}
	if env != "" {
		err := LoadEnv(env)
		CheckError(err, "loading dotenv file %s: %v", env)
	}
	ConnectDatabase()
	defer db.Close()
	defer tx.Commit()
	if fill {
		RunFill(dir, upto)
	} else if dry {
		RunDry(dir, upto, init)
	} else {
		RunMain(dir, upto, init)
	}
	PrintOK()
}
