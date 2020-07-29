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
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	// QueryMetaExists is the query to check that meta table exists
	QueryMetaExists = `
	SELECT count(*) FROM _stork;
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

// ParseCommandLine does what you think
func ParseCommandLine() (string, string, error) {
	env := flag.String("env", "", "Dotenv file to load")
	flag.Parse()
	dirs := flag.Args()
	dir := "."
	if len(dirs) > 1 {
		return "", "", fmt.Errorf("You can pass only one directory")
	}
	if len(dirs) == 1 {
		dir = dirs[0]
	}
	return *env, dir, nil
}

// LoadEnv loads environment in given file
func LoadEnv(filename string) error {
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
	return scripts, nil
}

// ConnectDatabase returns database connection
func ConnectDatabase() (*sql.DB, error) {
	source := os.Getenv("MYSQL_USERNAME") + ":" + os.Getenv("MYSQL_PASSWORD") +
		"@" + os.Getenv("MYSQL_HOSTNAME") + "/" + os.Getenv("MYSQL_DATABASE")
	return sql.Open("mysql", source)
}

// InitMetaTable initializes meta tables if necessary
func InitMetaTable() error {
	db, err := ConnectDatabase()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Query(QueryMetaExists)
	if err != nil {
		fmt.Println("Creating meta table")
		create, err := db.Query(QueryCreateMeta)
		if err != nil {
			return err
		}
		defer create.Close()
	}
	return nil
}

// ScriptPassed tells if given script already passed
func ScriptPassed(script string) (bool, error) {
	db, err := ConnectDatabase()
	if err != nil {
		return false, err
	}
	defer db.Close()
	var count int
	err = db.QueryRow(QueryScriptPassed, script).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// RunScript runs given script
func RunScript(dir, script string) error {
	fmt.Printf("Passing script %s\n", script)
	file, err := os.Open(filepath.Join(dir, script))
	if err != nil {
		return err
	}
	defer file.Close()
	source, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	db, err := ConnectDatabase()
	if err != nil {
		return err
	}
	result, err := db.Query(string(source))
	if err != nil {
		defer result.Close()
	}
	RecordResult(script, err)
	return nil
}

// RecordResult record script result in meta table
func RecordResult(script string, err error) error {
	db, err := ConnectDatabase()
	if err != nil {
		return err
	}
	var success bool
	var message string
	if err != nil {
		success = false
		message = err.Error()
	} else {
		success = true
		message = ""
	}
	result, err := db.Query(QueryRecordResult, script, success, message)
	if err != nil {
		return err
	}
	defer result.Close()
	return nil
}

func main() {
	env, dir, err := ParseCommandLine()
	if err != nil {
		fmt.Printf("Error parsing command line: %v\n", err)
		os.Exit(1)
	}
	if env != "" {
		err := LoadEnv(env)
		if err != nil {
			fmt.Printf("Error loading dotenv file %s: %v\n", env, err)
			os.Exit(1)
		}
	}
	scripts, err := ScriptsList(dir)
	if err != nil {
		fmt.Printf("Error getting scripts list: %v\n", err)
		os.Exit(1)
	}
	err = InitMetaTable()
	if err != nil {
		fmt.Printf("Error initializing meta tables: %v\n", err)
		os.Exit(1)
	}
	for _, script := range scripts {
		passed, err := ScriptPassed(script)
		if err != nil {
			fmt.Printf("Error determining if script %s passed: %v\n", script, err)
			os.Exit(1)
		}
		if !passed {
			err = RunScript(dir, script)
			if err != nil {
				fmt.Printf("Error running script %s: %v\n", script, err)
			}
		}
	}
}
