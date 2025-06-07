package astinfo

import "go/ast"

type Method struct {
	Function
	Receiver *Struct
}

// new
func NewMethod(funcDecl *ast.FuncDecl, goSource *Gosourse) *Method {
	return &Method{
		Function: *NewFunction(funcDecl, goSource),
	}
}
func (m *Method) Parse() error {
	// 方法体为空
	return nil
}
