package astbasic

import (
	"strings"
	"unicode"
)

// ToCamelCase converts snake_case to lowerCamelCase
// userId => UserID(go) => userId(json)
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
	if isGo && strings.HasSuffix(result, "Id") {
		//解决数据中没有使用驼峰，且Id的问题；
		result = result[:len(result)-2] + "ID"
	}
	return result
}

var abbrMap = map[string]string{
	"id":   "ID",
	"url":  "URL",
	"ip":   "IP",
	"uid":  "UID",
	"uuid": "UUID",
	"json": "JSON",
	"html": "HTML",
	"xml":  "XML",
	"db":   "DB",
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
