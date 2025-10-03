package tool

import (
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

type split struct {
	index   int
	content []byte
	result  []string
}

func (s *split) move(stop byte) {
	for s.index < len(s.content) && s.content[s.index] != stop {
		s.index++
	}
}
func (s *split) split() {
	begin := 0
	for ; s.index < len(s.content); s.index++ {
		a := s.content[s.index]
		switch a {
		case '"':
			s.index++
			s.move('"')
		case ' ', '\t', ';':
			if begin < s.index {
				s.result = append(s.result, string(s.content[begin:s.index]))
			}
			begin = s.index + 1
		}
	}
	if begin < s.index {
		s.result = append(s.result, string(s.content[begin:s.index]))
	}
}

func Fields(s string) []string {
	a := split{
		content: []byte(s),
	}
	a.split()
	return a.result
}
func ToPascalCase(s string, firstBig bool) string {
	if s == "_id" {
		return "ID"
	}
	var result strings.Builder
	capitalizeNext := firstBig
	for _, r := range s {
		if r == '_' || r == '-' {
			capitalizeNext = true
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if capitalizeNext {
				result.WriteRune(unicode.ToUpper(r))
				capitalizeNext = false
			} else {
				result.WriteRune(r)
			}
		}
	}
	return result.String()
}
