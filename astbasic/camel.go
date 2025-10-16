package astbasic

import (
	"strings"
	"unicode"
)

// ToCamelCase converts snake_case to lowerCamelCase
func ToCamelCase(s string, isGo bool) string {

	words := strings.Split(s, "_")
	if len(words) == 0 {
		return s
	}
	var result string
	if isGo {
		result = firstBig(words[0], isGo)
	} else {
		result = words[0]
	}
	for _, word := range words[1:] {
		result += firstBig(word, isGo)
	}
	return result
}

var abbrMap = map[string]string{
	"id":  "ID",
	"url": "URL",
}

func firstBig(word string, isGo bool) string {
	var result string
	if len(word) == 0 {
		return result
	}

	if isGo {
		if abbr, ok := abbrMap[strings.ToLower(word)]; ok {
			return abbr
		}
	}

	runes := []rune(word)
	result = string(unicode.ToUpper(runes[0])) + string(runes[1:])
	return result
}
