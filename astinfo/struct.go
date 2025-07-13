package astinfo

import "go/ast"

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
	StructName    string // 结构体名称
	Pkg           *Package
	genDecl       *ast.GenDecl
	astStructType *ast.StructType
	comment       structComment
	MethodManager
	Fields []*Field
	// TODO: 后续添加字段和方法解析
}

func (v *Struct) IsPointer() bool {
	return false
}

func (v *Struct) Name(genFile *GenedFile) string {
	if genFile.pkg == v.Pkg {
		return v.StructName
	}
	impt := genFile.getImport(v.Pkg)
	return impt.Name + "." + v.StructName
}

func (v *Struct) FullName() string {
	return v.Pkg.name + "." + v.StructName
}

// new
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
	v.Fields = parseFields(v.astStructType.Fields.List, v.Pkg)
	return nil
}
