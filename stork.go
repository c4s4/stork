package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
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
	fmt.Printf("dir: %s\n", dir)
}
