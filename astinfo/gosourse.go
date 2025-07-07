package astinfo

import (
	"fmt"
	"go/ast"
	"go/token"
)

type Gosourse struct {
	Path string
	pkg  *Package
	File *ast.File
}

// 解析全局变量和 struct，interface
func (g *Gosourse) getGenDeclParser(genDecl *ast.GenDecl) (parser Parser) {
	switch genDecl.Tok {
	case token.VAR:
	case token.TYPE:
		typeSpec := genDecl.Specs[0].(*ast.TypeSpec)
		// 仅关注结构体，暂时不考虑接口
		switch typeSpec.Type.(type) {
		case *ast.InterfaceType:

		case *ast.StructType:
			class := g.pkg.FindStruct(typeSpec.Name.Name)
			class.initGenDecl(genDecl)
			parser = class
		}
	}
	return
}

// 解析函数和方法
func (g *Gosourse) getFuncDeclParser(funcDecl *ast.FuncDecl) Parser {
	switch funcDecl.Recv {
	case nil:
		return NewFunctionParserHelper(funcDecl, g)
	default:
		return NewMethod(funcDecl, g)
	}
}
func (g *Gosourse) Parse() error {
	fmt.Printf("Parsing file: %s\n", g.Path)
	if g.pkg.name == "" {
		g.pkg.name = g.File.Name.Name
	} else if g.pkg.name != g.File.Name.Name {
		// 这里是有问题的，需要修改
		// 不报错了。原工程会报错
	}
	decls := g.File.Decls
	for i := 0; i < len(decls); i++ {
		var parser Parser
		switch decl := decls[i].(type) {
		case *ast.GenDecl:
			parser = g.getGenDeclParser(decl)
		case *ast.FuncDecl:
			parser = g.getFuncDeclParser(decl)
		}
		if parser != nil {
			parser.Parse()
		}
	}
	// 方法体为空
	return nil
}

func NewGosourse(file *ast.File, pkg *Package, path string) *Gosourse {
	return &Gosourse{
		File: file,
		pkg:  pkg,
		Path: path,
	}
}
