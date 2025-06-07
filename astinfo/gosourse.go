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
func (g *Gosourse) parseGenDecl(genDecl *ast.GenDecl) Parser {
	switch genDecl.Tok {
	case token.VAR:
	case token.TYPE:
	}
	return nil
}

// 解析函数和方法
func (g *Gosourse) parseFuncDecl(funcDecl *ast.FuncDecl) Parser {
	switch funcDecl.Recv {
	case nil:
		return NewFunction(funcDecl, g)
	default:
		return NewMethod(funcDecl, g)
	}
	return nil
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
			parser = g.parseGenDecl(decl)
		case *ast.FuncDecl:
			parser = g.parseFuncDecl(decl)
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
