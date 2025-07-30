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
type FieldBasic struct {
	Type    Typer // 实际可以为Struct，Interface， RawType
	Name    string
	Comment FieldComment
	// astRoot  *ast.Field
	GoSource *Gosourse //解析Filed时，其他type可能来源其他Package，此时需要Import内容来找到该包；

	astDoc     *ast.CommentGroup // associated documentation; or nil
	astNames   []*ast.Ident      // value names (len(Names) > 0)
	astType    ast.Expr          // value type; or nil
	astComment *ast.CommentGroup // line comments; or nil
}

// genVariableCode
func (f *FieldBasic) GenVariableCode(goGenerated *GenedFile, wire bool) string {
	if f.Type == nil {
		fmt.Printf("skip gen variable for field %s as type is nil in %s\n", f.Name, f.GoSource.Path)
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
				field.Tags[kv[0]] = strings.Trim(kv[1], "\"")
			}
		}
	}
}
func (field *FieldBasic) parseComment(fieldType *ast.CommentGroup) {
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
func (field *FieldBasic) Parse() error {
	fieldType := field.astType
	//field.Name="名字在调用本函数的外面解析，因为一个类型可能有多个名字，需要拆分为多个Field"
	field.ParseType(fieldType)
	field.parseComment(field.astComment)

	return nil
}

func findType(pkg *Package, typeName string) Typer {
	// 	res := pkg.GetTyper(typeName)
	// if res == nil {
	// 	fmt.Printf("find type %s failed\n", typeName)
	// }
	// return res
	typer := pkg.FindStruct(typeName)
	if typer == nil {
		typer1 := pkg.FindInterface(typeName)
		if typer1 != nil {
			return typer1
		}
	} else {
		return typer
	}
	return nil
}

// 在pkg内解析Type；
func (field *FieldBasic) parseType(typer *Typer, fieldType ast.Expr) error {
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
		var pointer Typer
		err = field.parseType(&pointer, fieldType.X)
		resultType = NewPointerType(pointer)
	case *ast.Ident:
		// 此时可能是
		// 原始类型； string
		// 同package的结构体，
		// field.Type =
		// 先检查原始类型；
		type1 := GetRawType(fieldType.Name)
		if type1 == nil {
			//再检查Struct类型；
			resultType = findType(field.GoSource.Pkg, fieldType.Name)
		} else {
			resultType = type1
		}
	case *ast.SelectorExpr:
		// 其他package的结构体，=》pkg1.Struct
		// field定义的selector，就只考虑pkg1
		pkgName := fieldType.X.(*ast.Ident).Name
		typeName := fieldType.Sel.Name
		pkgModePath := field.GoSource.Imports[pkgName]
		resultType = findType(GlobalProject.FindPackage(pkgModePath), typeName)
	default:
		fmt.Printf("unknown field type '%T' in '%s'\n", fieldType, field.GoSource.Path)
		return nil
	}
	//如果将来Typer需要全局唯一，此处可以先找到唯一值，再赋值给typer；
	*typer = resultType
	return err
}

func (field *FieldBasic) ParseType(fieldType ast.Expr) error {
	return field.parseType(&field.Type, fieldType)
}

type Field struct {
	FieldBasic
	Tags   map[string]string
	astTag *ast.BasicLit
}

func (field *Field) Parse() error {
	field.parseTag(field.astTag)
	return field.FieldBasic.Parse()
}

func NewField(root *ast.Field, source *Gosourse) *Field {
	return &Field{
		FieldBasic: FieldBasic{
			GoSource:   source,
			astDoc:     root.Doc,
			astNames:   root.Names,
			astType:    root.Type,
			astComment: root.Comment,
		},
		astTag: root.Tag,
		Tags:   make(map[string]string),
	}
}
func NewSimpleField(typer Typer, name string) *Field {
	return &Field{
		FieldBasic: FieldBasic{
			Name: name,
			Type: typer,
		},
	}
}

type VarFieldHelper struct {
	varField FieldBasic
	astRoot  *ast.ValueSpec
	goSource *Gosourse
}

type VarField = FieldBasic

func NewVarFieldHelper(root *ast.ValueSpec, source *Gosourse) *VarFieldHelper {
	return &VarFieldHelper{
		astRoot:  root,
		goSource: source,
	}
}

func (v *VarFieldHelper) Parse() error {
	var root = v.astRoot
	var field = FieldBasic{
		GoSource:   v.goSource,
		astDoc:     root.Doc,
		astNames:   root.Names,
		astType:    root.Type,
		astComment: root.Comment,
	}
	field.Parse()
	if len(root.Names) != 0 {
		for _, name := range root.Names {
			field1 := field
			field1.Name = name.Name
			v.goSource.Pkg.GlobalVar[name.Name] = &field1
		}
	}
	return nil
}
