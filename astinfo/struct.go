package astinfo

import (
	"go/ast"
	"strings"
)

// @goservlet prpc=xxx; servlet=xxx; servlet; prpc
// @goservlet type=xxx ;  prpc, servlet, websocket, restful,
// @goservlet group=xxx; 如果不存在，则跟type同名
type structComment struct {
	groupName  string
	serverType string // NONE, RpcStruct, ServletStruct·
	url        string // 服务的url, 对所有的方法都有效
}

func (comment *structComment) dealValuePair(key, value string) {
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
	}
}

// Struct 表示一个Go结构体的基本信息
type Struct struct {
	StructName    string    // 结构体名称
	goSource      *Gosourse //该变量在解析结构体时赋值，也就是说该变量不为空，则该结构体已经被解析；
	Pkg           *Package  //该变量肯定不为空，但是goSource不一定；
	genDecl       *ast.GenDecl
	astStructType *ast.StructType
	comment       structComment
	Fields        []*Field
	MethodManager
	// TODO: 后续添加字段和方法解析
}

func (v *Struct) Name(genFile *GenedFile) string {
	pkg := v.Pkg
	if genFile.pkg == pkg {
		return v.StructName
	}
	impt := genFile.getImport(pkg)
	return impt.Name + "." + v.StructName
}

func (v *Struct) FullName() string {
	return v.Pkg.name + "." + v.StructName
}
func (v *Struct) GenConstructCode(genFile *GenedFile) string {
	result := genFile.getImport(v.Pkg)
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
	sb.WriteString("}")

	return sb.String()
}

// 不一定每次newStruct时都会有goSrouce，所以此时只能传Pkg；
func NewStruct(name string, pkg *Package) *Struct {
	return &Struct{
		StructName: name,
		Pkg:        pkg,
	}
}

// initGenDecl
func (v *Struct) initGenDecl(genDecl *ast.GenDecl) {
	v.genDecl = genDecl
	v.astStructType = genDecl.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)
}

// 解析结构体的注释和字段
func (v *Struct) Parse() error {
	if err := v.ParseComment(); err != nil {
		return err
	}
	return v.ParseField()
}

// parseComment
func (class *Struct) ParseComment() error {
	parseComment(class.genDecl.Doc, &class.comment)
	return nil
}

// parseField
func (v *Struct) ParseField() error {
	// v.goSource在解析结构体时，被赋值，解析field也是在解析结构体时，所以v.goSource不为空
	v.Fields = parseFields(v.astStructType.Fields.List, v.goSource)
	return nil
}
