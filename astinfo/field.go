package astinfo

import (
	"fmt"
	"go/ast"
)

type FieldComment struct {
}

// 变量名和变量类型的定义
// 用于函数的参数和返回值，struct的属性；
type Field struct {
	Type     Typer // 实际可以为Struct，Interface， RawType
	Name     string
	Comment  FieldComment
	astRoot  *ast.Field
	goSource *Gosourse //解析Filed时，其他type可能来源其他Package，此时需要Import内容来找到该包；
}

// genVariableCode
func (f *Field) GenVariableCode(goGenerated *GenedFile) string {
	variable := Variable{
		Type: f.Type,
		Name: f.Name,
	}
	return variable.Generate(goGenerated)
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

// 在pkg内解析Type；
func (field *Field) parseType(typer *Typer, fieldType ast.Expr) error {
	var resultType Typer
	var err error
	switch fieldType := fieldType.(type) {
	case *ast.ArrayType:
		// 内置array类型
		// field的pkg指向原始类型；
		// field的class只想ArrayType;
		// ArrayType中的pkg，typeName，class指向具体的类型
		array := ArrayType{}
		resultType = &array
		err = field.parseType(&array.Typer, fieldType.Elt)
	case *ast.StarExpr:
		pointer := PinterType{}
		resultType = &pointer
		if p, ok := pointer.Typer.(*PinterType); ok {
			pointer.Depth = p.Depth + 1
		} else {
			pointer.Depth = 1
		}
		err = field.parseType(&pointer.Typer, fieldType.X)
	case *ast.Ident:
		// 此时可能是
		// 原始类型； string
		// 同package的结构体，
		// field.Type =
		// 先检查原始类型；
		type1 := GetRawType(fieldType.Name)
		if type1 == nil {
			//再检查Struct类型；
			resultType = findType(field.goSource.pkg, fieldType.Name)
		} else {
			resultType = type1
		}
	case *ast.SelectorExpr:
		// 其他package的结构体，=》pkg1.Struct
		// field定义的selector，就只考虑pkg1
		pkgName := fieldType.X.(*ast.Ident).Name
		typeName := fieldType.Sel.Name
		pkgModePath := field.goSource.Imports[pkgName]
		resultType = findType(GlobalProject.FindPackage(pkgModePath), typeName)
	default:
		fmt.Printf("unknown field type '%T'\n", fieldType)
		return nil
	}
	//如果将来Typer需要全局唯一，此处可以先找到唯一值，再赋值给typer；
	*typer = resultType
	return err
}

func (field *Field) ParseType(fieldType ast.Expr) error {
	return field.parseType(&field.Type, fieldType)
}
