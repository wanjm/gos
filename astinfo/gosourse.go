package astinfo

import "go/ast"

type Gosourse struct {
	Path    string
	Package *Package
	File    *ast.File
}

func (g *Gosourse) Parse() error {
	// 方法体为空
	return nil
}

func NewGosourse(file *ast.File, pkg *Package) *Gosourse {
	return &Gosourse{
		File:    file,
		Package: pkg,
	}
}
