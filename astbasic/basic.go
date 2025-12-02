package astbasic

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

func FirstLower(word string) string {
	return string(unicode.ToLower([]rune(word)[0])) + word[1:]
}
func Capitalize(s string) string {
	if s == "" {
		return s
	}
	// 将字符串转换为rune切片处理Unicode字符
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

func Fields(s string, delimiter ...byte) []string {
	if len(delimiter) == 0 {
		delimiter = []byte{'\t', ' ', ';'}
	}

	var result []string
	start := 0
	inQuote := false

	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			inQuote = !inQuote
		} else if !inQuote {
			isDelim := slices.Contains(delimiter, s[i])
			if isDelim {
				if i > start {
					result = append(result, s[start:i])
				}
				start = i + 1
			}
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
func ConvertGoTag(tag string) string {
	// 处理带转义符的情况，如"json:\"hello\""
	if strings.HasPrefix(tag, "\"") && strings.HasSuffix(tag, "\"") {
		// 去除首尾的引号
		trimmed := tag[1 : len(tag)-1]
		// 替换转义的引号为普通引号
		return strings.ReplaceAll(trimmed, "\\\"", "\"")
	}

	// 处理原始tag情况，如`json:"hello"`
	if strings.HasPrefix(tag, "`") && strings.HasSuffix(tag, "`") {
		return tag[1 : len(tag)-1]
	}

	// 对于其他情况，直接返回原字符串
	return tag
}
