package astinfo

import (
	"fmt"
	"go/ast"
	"strings"
)

type FieldComment struct {
}

// 变量名和变量类型的定义
// 用于函数的参数和返回值，struct的属性；
type Field struct {
	Type         Typer // 实际可以为Struct，Interface， RawType
	Name         string
	isPointer    bool
	pointerCount int
	Comment      FieldComment
	astRoot      *ast.Field
	pkg          *Package
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
// 此函数仅解析结构，然后在外面解析名字，拆分为多个Field
func (field *Field) Parse() error {
	fieldType := field.astRoot.Type
	// var modeName, structName strings
	// 内置slice类型；
	field.ParseType(fieldType)
	return nil
}
func findType(pkg *Package, typeName string) Typer {
	return pkg.FindStruct(typeName)
}
func (field *Field) ParseType(fieldType ast.Expr) error {
	switch fieldType := fieldType.(type) {
	case *ast.ArrayType:
		// 内置array类型
		// field的pkg指向原始类型；
		// field的class只想ArrayType;
		// ArrayType中的pkg，typeName，class指向具体的类型
		field.Type = &ArrayType{}
		return nil
	case *ast.StarExpr:
		field.isPointer = true
		field.pointerCount++
		return field.ParseType(fieldType.X)
	case *ast.Ident:
		// 此时可能是
		// 原始类型； string
		// 同package的结构体，
		// field.Type =
		// 先检查原始类型；
		type1 := GetRawType(fieldType.Name)
		if type1 == nil {
			//再检查Struct类型；
			field.Type = findType(field.pkg, fieldType.Name)
		} else {
			field.Type = type1
		}
		return nil
	case *ast.SelectorExpr:
		// 其他package的结构体，=》pkg1.Struct
		// field定义的selector，就只考虑pkg1
		pkgName := fieldType.X.(*ast.Ident).Name
		typeName := fieldType.Sel.Name
		field.Type = findType(GlobalProject.FindPackage(pkgName), typeName)
		return nil
	default:
		fmt.Printf("unknown field type '%T'\n", fieldType)
	}
	return nil
}
