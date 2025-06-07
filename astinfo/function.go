package astinfo

import "go/ast"

type Function struct {
	funcDecl *ast.FuncDecl
	goSource *Gosourse
}

// create
func NewFunction(funcDecl *ast.FuncDecl, goSource *Gosourse) *Function {
	return &Function{
		funcDecl: funcDecl,
		goSource: goSource,
	}
}
func (f *Function) Parse() error {
	// 方法体为空
	return nil
}
