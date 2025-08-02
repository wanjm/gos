package astinfo

import (
	"fmt"
	"go/ast"
	"strings"
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
		config.Type = strings.Trim(value, `"`)
	default:
		fmt.Printf("unkonw key value pair => key=%s,value=%s\n", key, value)
	}
}

type Interface struct {
	Comment       interfaceComments
	GoSource      *Gosourse
	InterfaceName string
	// Pkg           *Package

	genDecl *ast.GenDecl
	astRoot *ast.InterfaceType
	Methods []*InterfaceField
}

func NewInterface(name string, goSource *Gosourse, genDecl *ast.GenDecl) *Interface {
	iface := &Interface{
		InterfaceName: name,
		GoSource:      goSource,
	}
	pkg := goSource.Pkg
	// pkg.Interfaces[name] = iface
	pkg.Types[name] = iface
	iface.initGenDecl(genDecl)
	return iface
}
func (i *Interface) Parse() error {
	parseComment(i.genDecl.Doc, &i.Comment)
	// Type 为空表示不是client interface，跳过处理
	if i.Comment.Type == "" {
		return nil
	}
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

// RefName returns the type name with package prefix if needed
func (i *Interface) RefName(genFile *GenedFile) string {
	pkg := i.GoSource.Pkg
	if genFile == nil || genFile.pkg == pkg {
		return i.InterfaceName
	}

	impt := genFile.GetImport(pkg)
	return impt.Name + "." + i.InterfaceName
}

// IDName returns the full name of the interface with package path
func (i *Interface) IDName() string {
	return i.GoSource.Pkg.Module + "." + i.InterfaceName
}

// GenConstructCode generates code to construct an instance of the interface
func (i *Interface) GenConstructCode(genFile *GenedFile, wire bool) string {
	// Interfaces cannot be constructed directly, they need to be implemented by structs
	// Return nil value for interface
	return ""
}
