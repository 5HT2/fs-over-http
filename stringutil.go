package main

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

func JoinStr(str string, suffix string) string {
	strArr := []string{str, suffix}
	return strings.Join(strArr, "")
}

func TrimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

func Grammar(amount int, singular string, multiple string) string {
	if amount != 1 {
		return strings.Join([]string{strconv.Itoa(amount), multiple}, " ")
	} else {
		return strings.Join([]string{strconv.Itoa(amount), singular}, " ")
	}
}

// LastByte returns the last byte of a byte array
// and also returns the string as a byte array
func LastByte(s string) (byte, []byte) {
	b := []byte(s)
	return b[len(b)-1], b
}
