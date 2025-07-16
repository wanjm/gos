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

// 解析method
// 首先解析receiver；找到自己所属的Struct；
func (m *Method) Parse() error {
	if err := m.parseReceiver(); err != nil {
		return err
	}
	m.Function.Parse()
	return nil
}

func (m *Method) parseReceiver() error {
	// 方法体为空
	recvType := m.funcDecl.Recv.List[0].Type
	var nameIndent *ast.Ident
	if starExpr, ok := recvType.(*ast.StarExpr); ok {
		nameIndent = starExpr.X.(*ast.Ident)
	} else {
		nameIndent = recvType.(*ast.Ident)
	}
	// 由于代码的位置关系，这一步不一定会找到，所以自己创建了。
	receiver := m.pkg.FindStruct(nameIndent.Name)
	m.Receiver = receiver
	receiver.MethodManager.AddCallable(m)
	return nil
}
