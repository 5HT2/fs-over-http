package main

import (
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// AddLastRune will check the last rune of a string,
// if the last rune does not match, it will append the rune,
// and return the fixed string
func AddLastRune(str string, rune rune) (fixed string) {
	fixed = str

	r, _ := utf8.DecodeLastRuneInString(str)
	if r != rune {
		fixed = JoinStr(str, string(rune))
	}

	return fixed
}

func RemoveLastRune(str string, rune rune) (fixed string) {
	fixed = str

	r, _ := utf8.DecodeLastRuneInString(str)
	if r == rune {
		fixed = TrimLastRune(str)
	}

	return fixed
}

// TrimFirstRune will remove the first rune in a string
func TrimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

// TrimLastRune will remove the last rune in a string
func TrimLastRune(s string) string {
	s = s[:len(s)-1]
	return s
}

func JoinStr(str string, suffix string) string {
	strArr := []string{str, suffix}
	return strings.Join(strArr, "")
}

func Grammar(amount int, singular string, multiple string) string {
	if amount != 1 {
		return strings.Join([]string{strconv.Itoa(amount), multiple}, " ")
	} else {
		return strings.Join([]string{strconv.Itoa(amount), singular}, " ")
	}
}

// TODO: This breaks if the string is empty

// LastByte returns the last byte of a byte array
// and also returns the string as a byte array
func LastByte(s string) (byte, []byte) {
	b := []byte(s)
	return b[len(b)-1], b
}

// TODO: This breaks if the slice contains blank spaces
func Contains(s []string, term string) bool {
	i := sort.SearchStrings(s, term)
	return i < len(s) && s[i] == term
}
