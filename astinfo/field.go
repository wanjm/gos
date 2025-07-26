package astinfo

import (
	"fmt"
	"go/ast"
	"strings"
)

type FieldComment struct {
	defaultValue string //记录该属性的默认值，在struct的field中有使用；
	// isRequired   bool   //记录该字段是否必须赋值，区别于gin的默认处理方法，必传表示在报文中必须存在
	validString string //校验变量是否符合要求的代码； $>10 && $<11
	comment     string
}

func (comment *FieldComment) dealValuePair(key, value string) {
	switch key {
	case "default":
		comment.defaultValue = value
	case "valid":
		comment.validString = value
	default:
		comment.comment = key
	}
}

// 变量名和变量类型的定义
// 用于函数的参数和返回值，struct的属性；
type Field struct {
	Type     Typer // 实际可以为Struct，Interface， RawType
	Name     string
	Comment  FieldComment
	astRoot  *ast.Field
	goSource *Gosourse //解析Filed时，其他type可能来源其他Package，此时需要Import内容来找到该包；
	Tags     map[string]string
}

func NewField(root *ast.Field, source *Gosourse) *Field {
	return &Field{
		astRoot:  root,
		goSource: source,
		Tags:     make(map[string]string),
	}
}

// genVariableCode
func (f *Field) GenVariableCode(goGenerated *GenedFile, wire bool) string {
	if f.Type == nil {
		fmt.Printf("skip gen variable for field %s as type is nil in %s\n", f.Name, f.goSource.Path)
		return ""
	}
	variable := Variable{
		Type: f.Type,
		Name: f.Name,
		Wire: wire,
	}
	return variable.Generate(goGenerated)
}

func (field *Field) parseTag(fieldType *ast.BasicLit) {
	if fieldType != nil {
		tag := fieldType.Value
		tag = strings.Trim(tag, string(tag[0]))
		// if strings.Contains(tag, "wire") {
		// 	fmt.Print("hello")
		// }
		tagList := Fields(tag)
		for _, tag := range tagList {
			kv := strings.Split(tag, ":")
			if len(kv) == 2 {
				field.Tags[kv[0]] = kv[1]
			}
		}
	}
}
func (field *Field) parseComment(fieldType *ast.CommentGroup) {
	if fieldType == nil || len(fieldType.List) <= 0 {
		return
	}
	content := strings.Trim(fieldType.List[0].Text, " /")
	parseValidComment(content, &field.Comment)
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
	field.parseComment(field.astRoot.Comment)
	field.parseTag(field.astRoot.Tag)
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
		pointer := PointerType{}
		resultType = &pointer
		if p, ok := pointer.Typer.(*PointerType); ok {
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
			resultType = findType(field.goSource.Pkg, fieldType.Name)
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
		fmt.Printf("unknown field type '%T' in '%s'\n", fieldType, field.goSource.Path)
		return nil
	}
	//如果将来Typer需要全局唯一，此处可以先找到唯一值，再赋值给typer；
	*typer = resultType
	return err
}

func (field *Field) ParseType(fieldType ast.Expr) error {
	return field.parseType(&field.Type, fieldType)
}
