package astinfo

import (
	"fmt"
	"go/ast"
)

type Function struct {
	funcDecl        *ast.FuncDecl
	goSource        *Gosourse
	functionManager *FunctionManager
}

// create
func NewFunction(funcDecl *ast.FuncDecl, goSource *Gosourse) *Function {
	return &Function{
		funcDecl: funcDecl,
		goSource: goSource,
	}
}
func (f *Function) Parse() error {
	if f.functionManager == nil {
		fmt.Printf("functionManager should be initialized")
	}
	// 方法体为空
	return nil
}
