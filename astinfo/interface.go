package astinfo

import (
	"fmt"
	"go/ast"
)

type interfaceComments struct {
	Host string
	Type string
}

func (config *interfaceComments) dealValuePair(key, value string) {
	switch key {
	case Host:
		config.Host = value
	case Type:
		config.Type = value
	default:
		fmt.Printf("unkonw key value pair => key=%s,value=%s\n", key, value)
	}
}

type Interface struct {
	Comment       interfaceComments
	GoSource      *Gosourse
	InterfaceName string
	Pkg           *Package

	genDecl *ast.GenDecl
	astRoot *ast.InterfaceType
	Methods []*InterfaceField
}

func NewInterface(name string, pkg *Package) *Interface {
	return &Interface{
		InterfaceName: name,
		Pkg:           pkg,
	}
}
func (i *Interface) Parse() error {
	parseComment(i.genDecl.Doc, &i.Comment)
	// 方法体为空
	i.parseBody()
	return nil
}

func (i *Interface) parseBody() error {
	// 方法体为空
	for _, method := range i.astRoot.Methods.List {
		methodField := NewInterfaceField(method, i.GoSource)
		methodField.Parse()
		i.Methods = append(i.Methods, methodField)
	}
	return nil
}

func (v *Interface) initGenDecl(genDecl *ast.GenDecl) {
	v.genDecl = genDecl
	v.astRoot = genDecl.Specs[0].(*ast.TypeSpec).Type.(*ast.InterfaceType)
}
