package astinfo

import (
	"fmt"
	"go/ast"
)

type Function struct {
	funcDecl        *ast.FuncDecl
	goSource        *Gosourse
	functionManager *FunctionManager //function指向pkg的functionManager，method指向自己recevicer(struct)的functionManager
}

// create
func NewFunction(funcDecl *ast.FuncDecl, goSource *Gosourse) *Function {
	return &Function{
		funcDecl: funcDecl,
		goSource: goSource,
	}
}

// 解析自己，并把自己添加到对应的functionManager中；
func (f *Function) Parse() error {
	if f.functionManager == nil {
		fmt.Printf("functionManager should be initialized")
	}
	// 方法体为空
	return nil
}
