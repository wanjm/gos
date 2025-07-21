package astinfo

import (
	"unicode"
	"unicode/utf8"
)

type Parser interface {
	Parse() error
}

type CodeGenerator interface {
	// 需要GoGenerated，作为参数的主要原因是，生成过程中需要向代码中添加import；

	Generate(goGenerated *GenedFile) error
}

type Import struct {
	Name string
	Path string
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	// 将字符串转换为rune切片处理Unicode字符
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}
