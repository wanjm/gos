package astinfo

import (
	"go/ast"
	"strings"
)

type FieldComment struct {
}

// 变量名和变量类型的定义
// 用于函数的参数和返回值，struct的属性；
type Field struct {
	Type      Typer // 实际可以为Struct，Interface， RawType
	Name      string
	isPointer bool
	Comment   FieldComment
	astRoot   ast.Expr
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

// Parse() error
// name type;
// name map
// name []arrays
func (field *Field) Parse() error {
	fieldType := field.astRoot
	// var modeName, structName strings
	// 内置slice类型；
	if arrayType, ok := fieldType.(*ast.ArrayType); ok {
		// 内置array类型
		// field的pkg指向原始类型；
		// field的class只想ArrayType;
		// ArrayType中的pkg，typeName，class指向具体的类型
		_ = arrayType
		field.Type = &ArrayType{}
		return nil
	}
	if innerType, ok := fieldType.(*ast.StarExpr); ok {
		field.isPointer = true
		fieldType = innerType.X
	}

	return nil
}
