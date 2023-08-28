package utils

import (
	"bufio"
	"strings"
)

// @ConvMultiStrToArrayStr - конвертируем многостроковый string в срез из строк
func ConvMultiStrToArrayStr(p string) []string {
	return makeStr(p)
}

// Make mutiline text to []string
func makeStr(mlstring string) []string {
	lineCount := lineCount(mlstring)

	resultArray := make([]string, lineCount)
	var i int

	scanner := bufio.NewScanner(strings.NewReader(mlstring))
	for scanner.Scan() {
		resultArray[i] = scanner.Text()
		i++
	}

	return resultArray
}

func lineCount(mlstring string) int {
	var lineCount int
	for i := range mlstring {
		if mlstring[i] == 10 {
			lineCount++
		}
	}
	return lineCount + 1
}
