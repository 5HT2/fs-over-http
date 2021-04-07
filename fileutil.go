package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func ReadFileUnsafe(file string, removeNewline bool) string {
	content, err := ReadFile(file)

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

func ReadFile(file string) (string, error) {
	dat, err := ioutil.ReadFile(file)
	return string(dat), err
}

func WriteToFile(file string, content string) error {
	data := []byte(content)
	err := ioutil.WriteFile(file, data, 0600)

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
