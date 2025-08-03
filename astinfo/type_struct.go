package astinfo

import (
	"go/ast"
	"strings"

	"github.com/go-openapi/spec"
)

// @goservlet prpc=xxx; servlet=xxx; servlet; prpc
// @goservlet type=xxx ;  prpc, servlet, websocket, restful,
// @goservlet group=xxx; 如果不存在，则跟type同名
type structComment struct {
	groupName  string
	serverType string // NONE, RpcStruct, ServletStruct·
	url        string // 服务的url, 对所有的方法都有效
	AutoGen    bool
}

func (comment *structComment) dealValuePair(key, value string) {
	comment.AutoGen = true
	switch key {
	case Prpc:
		comment.serverType = Prpc
		if len(value) == 0 {
			comment.groupName = Prpc
		} else {
			comment.groupName = value
		}
	case Servlet:
		comment.serverType = Servlet
		if len(value) == 0 {
			comment.groupName = Servlet
		} else {
			comment.groupName = value
		}
	case Group:
		comment.groupName = value
	case Type:
		comment.serverType = value
		if len(comment.groupName) == 0 {
			comment.groupName = comment.serverType
		}
	case Url:
		comment.url = value
	case AutoGen:
		comment.AutoGen = true
	}
}

// Struct 表示一个Go结构体的基本信息
type Struct struct {
	StructName string    // 结构体名称
	goSource   *Gosourse //该变量在解析结构体时赋值，也就是说该变量不为空，则该结构体已经被解析；
	// Pkg        *Package  //该变量肯定不为空，但是goSource不一定；
	astRoot *ast.TypeSpec
	// genDecl       *ast.GenDecl
	// astStructType *ast.StructType
	comment  structComment
	Fields   []*Field
	FieldMap map[string]*Field
	MethodManager
	// TODO: 后续添加字段和方法解析
	ref *spec.Ref
}

func (v *Struct) RefName(genFile *GenedFile) string {
	pkg := v.goSource.Pkg
	if genFile == nil || genFile.pkg == pkg {
		return v.StructName
	}
	impt := genFile.GetImport(pkg)
	return impt.Name + "." + v.StructName
}

func (v *Struct) IDName() string {
	return v.goSource.Pkg.Module + "." + v.StructName
}

func needWire(field *Field) bool {
	if IsRawType(field.Type) {
		return false
	}
	if field.Name == "" || (field.Name[0] <= 'z' && field.Name[0] >= 'a') {
		return false
	}

	if field.Tags["wire"] == `-` {
		return false
	}
	return true
}

// 这里有两种情况。如果定义一个统一的wire，需要考虑一下；
// 1. 自动注入，dal的情况；
// 2. request请求的自动创建；
// 3. 系统原始类型；
// 4. 嵌套的其他结构体；对于嵌套的结构体，在注入时，是必须要找到的；
// 5. 初步考虑可以将wire变量定义为必须注入内容结构体变量；
// wire为true表示必须绑定结构体等；
func (v *Struct) GenConstructCode(genFile *GenedFile, wire bool) string {
	result := genFile.GetImport(v.goSource.Pkg)
	var sb strings.Builder
	if result.Name != "" {
		sb.WriteString(result.Name)
		sb.WriteString(".")
	}
	sb.WriteString(v.StructName + "{\n")
	//结尾不能有\n,否则后续代码不好写，有语法错误；如：最后两行会有语法错误
	// getAddr(Strurct{
	//}
	//)
	// 1. 有default，则wire；
	// 2. wire为ture，且不是简单结构体（needWire），则寻找值去绑定；
	for _, field := range v.Fields {
		v, ok := field.Tags["default"]
		if ok || (needWire(field) && wire) {
			sb.WriteString(field.Name + ":")
			if ok {
				//此处需要考虑default为字符串等各种情况；
				sb.WriteString(v)
			} else {
				sb.WriteString(field.GenVariableCode(genFile, wire))
			}
			sb.WriteString(",\n")
		}
	}
	sb.WriteString("}")

	return sb.String()
}

// RequiredFields 返回结构体自己的字段，过滤掉原始类型或wire标记为"-"的字段
func (v *Struct) RequiredFields() []*Field {
	var requiredFields []*Field
	for _, field := range v.Fields {
		// 过滤原始类型
		if needWire(field) {
			requiredFields = append(requiredFields, field)
		}
	}
	return requiredFields
}

// GeneredFields 返回结构体自己
func (v *Struct) GeneredFields() []*Field {
	// 创建一个表示结构体自身的变量，且是指针格式；
	field := NewSimpleField(NewPointerType(v), "")
	return []*Field{field}
}

// GenerateDependcyCode 生成创建结构体对象的代码
func (v *Struct) GenerateDependcyCode(goGenerated *GenedFile) string {
	a := v.GeneredFields()[0]
	return a.Type.GenConstructCode(goGenerated, true)
}

// 不一定每次newStruct时都会有goSrouce，所以此时只能传Pkg；
// func NewStruct(name string, pkg *Package) *Struct {
// 	class := &Struct{
// 		StructName: name,
// 		Pkg:        pkg,
// 	}
// 	pkg.Structs[name] = class
// 	pkg.Types[name] = class
// 	return class
// }

func NewStruct(goSource *Gosourse, astRoot *ast.TypeSpec) *Struct {
	name := astRoot.Name.Name
	iface := &Struct{
		StructName: name,
		goSource:   goSource,
		astRoot:    astRoot,
	}
	pkg := goSource.Pkg
	pkg.Structs[name] = iface
	pkg.Types[name] = iface
	return iface
}

// initGenDecl
// func (v *Struct) initGenDecl(genDecl *ast.GenDecl, structType *ast.StructType) {
// 	v.genDecl = genDecl
// 	v.astStructType = structType
// }

// 解析结构体的注释和字段
func (v *Struct) Parse() error {
	if err := v.ParseComment(); err != nil {
		return err
	}
	return v.ParseField()
}

// parseComment
func (class *Struct) ParseComment() error {
	parseComment(class.astRoot.Doc, &class.comment)
	return nil
}

// parseField
func (v *Struct) ParseField() error {
	// v.goSource在解析结构体时，被赋值，解析field也是在解析结构体时，所以v.goSource不为空
	v.Fields = parseFields(v.astRoot.Type.(*ast.StructType).Fields.List, v.goSource)
	v.FieldMap = make(map[string]*Field)
	for _, field := range v.Fields {
		v.FieldMap[field.Name] = field
	}
	return nil
}
