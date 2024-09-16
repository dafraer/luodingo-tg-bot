package bot

import (
	"strconv"
	"strings"
	"unicode"
)

func containsDigit(s string) bool {
	for _, v := range s {
		if !unicode.IsDigit(v) {
			return true
		}
	}
	return false
}

func parseFlipCallback(s string) (string, int, error) {
	str := strings.Split(s, " ")
	n, err := strconv.Atoi(str[1])
	if err != nil {
		return "", 0, err
	}
	return str[0], n, nil
}
