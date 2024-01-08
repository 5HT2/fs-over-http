package main

import (
	"bufio"
	"encoding/json"
	"io/fs"
	"log"
	"os"
)

func ReadFileUnsafe(file string, removeNewline bool) string {
	b, err := os.ReadFile(file)
	content := string(b)

	if err != nil {
		log.Printf("- Failed to read '%s'", file)
		return content
	}

	lastByte, cb := LastByte(content)

	if removeNewline && lastByte == 10 {
		cb = cb[:len(cb)-1]
		return string(cb)
	}

	return content
}

func ReadUserTokens() map[string]UserToken {
	dat, err := os.ReadFile("user_tokens.json")
	if err != nil {
		return map[string]UserToken{}
	}

	var tokens map[string]UserToken
	if err := json.Unmarshal(dat, &tokens); err != nil || tokens == nil {
		return map[string]UserToken{}
	}

	return tokens
}

func WriteToFile(file string, content string) error {
	data := []byte(content)
	err := os.WriteFile(file, data, 0600)

	if err != nil {
		log.Printf("- Failed to read '%s'", file)
	}

	return err
}

func IsDirectory(path string) (bool, error) {
	fi, err := os.Stat(path)

	if err != nil {
		return false, err
	}

	if fi.Mode().IsDir() {
		return true, nil
	} else {
		return false, nil
	}
}

// ReadLines reads a whole file into memory
// and returns a slice of its lines.
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// ReadNonEmptyLines reads non-empty lines
// using ReadLines and a filter.
func ReadNonEmptyLines(file string, prefix string) (ret []string) {
	lines, err := ReadLines(file)

	if err != nil {
		return []string{}
	}

	for _, s := range lines {
		lastByte, _ := LastByte(s)

		// If not a blank line
		if len(s) > 1 || lastByte != 10 {
			ret = append(ret, prefix+s)
		}
	}

	return ret
}

func SafeMkdir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, ownerPerm)

		if err != nil {
			log.Fatalf("- Error making '%s' - %v", dir, err)
		}
	}
}

func Filter(ss []fs.FileInfo, test func(fs.FileInfo) bool) (ret []fs.FileInfo) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
