package astinfo

import "go/ast"

// /@goservlet prpc=xxx; servlet=xxx; servle; prpc
type structComment struct {
	groupName  string
	serverType int // NONE, RpcStruct, ServletStruct
}

func (comment *structComment) dealValuePair(key, value string) {
	switch key {
	case Prpc:
		comment.serverType = PRPC
		if len(value) == 0 {
			comment.groupName = Prpc
		} else {
			comment.groupName = value
		}
	case Servlet:
		comment.serverType = SERVLET
		if len(value) == 0 {
			comment.groupName = Servlet
		} else {
			comment.groupName = value
		}
	}
}

// Struct 表示一个Go结构体的基本信息
type Struct struct {
	Name          string // 结构体名称
	Pkg           *Package
	genDecl       *ast.GenDecl
	astStructType *ast.StructType
	comment       structComment
	// TODO: 后续添加字段和方法解析
}

// new
func NewStruct(name string, pkg *Package) *Struct {
	return &Struct{
		Name: name,
		Pkg:  pkg,
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
	// 方法体为空
	parseComment(class.genDecl.Doc, &class.comment)
	// for _, servlet := range class.servlets {
	// 	if servlet.comment.serverName == "" {
	// 		servlet.comment.serverName = class.comment.groupName
	// 	}
	// }
	// if class.comment.serverType != NOUSAGE {
	// 	class.Package.Project.addServer(class.comment.groupName)
	// }
	return nil
}

// parseField
func (v *Struct) ParseField() error {
	// 方法体为空
	return nil
}
