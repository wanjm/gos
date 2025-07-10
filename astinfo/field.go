package astinfo

import "strings"

type FieldComment struct {
}

type Typer interface {
	IsPointer() bool
	Name() string
}

// 变量名和变量类型的定义
// 用于函数的参数和返回值，struct的属性；
type Field struct {
	Type      Typer // 实际可以为Struct，Interface， RawType
	Name      string
	isPointer bool
	Comment   FieldComment
}

// genVariableCode
func (f *Field) GenVariableCode(goGenerated *GenedFile) string {
	var code strings.Builder
	code.WriteString(f.Name)
	if f.Type.IsPointer() {
		code.WriteString(" *")
	}
	code.WriteString(" ")
	code.WriteString(f.Type.Name())
	return code.String()
}
