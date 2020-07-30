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
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	// QueryMetaExists is the query to check that meta table exists
	QueryMetaExists = `
	SELECT count(*) FROM _stork
	`
	// QueryEraseMeta is the query to erase meta table
	QueryEraseMeta = `
	DROP TABLE IF EXISTS _stork
	`
	// QueryCreateMeta is the query to create meta table
	QueryCreateMeta = `
	CREATE TABLE _stork (
		id INTEGER NOT NULL AUTO_INCREMENT,
		script TEXT NOT NULL,
		date DATETIME DEFAULT CURRENT_TIMESTAMP,
		success BOOLEAN NOT NULL,
		error TEXT,
		PRIMARY KEY (id)
	)`
	// QueryScriptPassed is the query to determine if a script already passed
	QueryScriptPassed = `
	SELECT count(*)
	  FROM _stork
	 WHERE script = ?
	   AND success = 1
	`
	// QueryRecordResult is the query to record script passing
	QueryRecordResult = `
	INSERT INTO _stork (script, success, error)
	VALUES (?, ?, ?)
	`
)

var db *sql.DB
var mute bool
var white bool

// ParseCommandLine does what you think
func ParseCommandLine() (string, string, bool, bool, bool, error) {
	env := flag.String("env", "", "Dotenv file to load")
	init := flag.Bool("init", false, "Run all scripts")
	mute := flag.Bool("mute", false, "Don't print logs")
	white := flag.Bool("white", false, "Don't print color")
	flag.Parse()
	dirs := flag.Args()
	dir := "."
	if len(dirs) > 1 {
		return "", "", false, false, false, fmt.Errorf("You can pass only one directory")
	}
	if len(dirs) == 1 {
		dir = dirs[0]
	}
	return *env, dir, *init, *mute, *white, nil
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
		"@tcp(" + os.Getenv("MYSQL_HOSTNAME") + ")/" + os.Getenv("MYSQL_DATABASE")
	db, err := sql.Open("mysql", source)
	if err != nil {
		Error("connecting database: %v", err)
	}
	return db
}

// EraseMetaTable initializes meta tables if necessary
func EraseMetaTable() error {
	Print("Erasing meta table")
	_, err := db.Exec(QueryEraseMeta)
	return err
}

// CreateMetaTable initializes meta tables if necessary
func CreateMetaTable() error {
	_, err := db.Exec(QueryMetaExists)
	if err != nil {
		Print("Creating meta table")
		_, err := db.Exec(QueryCreateMeta)
		if err != nil {
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
		if strings.HasSuffix(strings.ToLower(file.Name()), ".sql") {
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
			_, err := db.Exec(query)
			if err != nil {
				tx.Rollback()
				RecordResult(script, err)
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
	_, err = db.Exec(QueryRecordResult, script, success, message)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	env, dir, init, isMute, isWhite, err := ParseCommandLine()
	mute = isMute
	white = isWhite
	if err != nil {
		Error("parsing command line: %v", err)
	}
	if env != "" {
		err := LoadEnv(env)
		if err != nil {
			Error("loading dotenv file %s: %v", env, err)
		}
	}
	db = ConnectDatabase()
	defer db.Close()
	if init {
		err = EraseMetaTable()
		if err != nil {
			Error("erasing meta table: %v", err)
		}
	}
	scripts, err := ScriptsList(dir)
	if err != nil {
		Error("getting scripts list: %v", err)
	}
	err = CreateMetaTable()
	if err != nil {
		Error("initializing meta tables: %v", err)
	}
	for _, script := range scripts {
		passed, err := ScriptPassed(script)
		if err != nil {
			Error("determining if script %s passed: %v", script, err)
		}
		if !passed {
			err = RunScript(dir, script)
			if err != nil {
				Error("running script %s: %v", script, err)
			}
		} else {
			Print("Skipping script %s", script)
		}
	}
	PrintOK()
}
