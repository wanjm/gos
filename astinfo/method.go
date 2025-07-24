package astinfo

import (
	"fmt"
	"go/ast"
)

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
	m.Function.Parse()
	if err := m.parseReceiver(); err != nil {
		return err
	}
	return nil
}

func (m *Method) parseReceiver() error {
	// 方法体为空
	recvType := m.funcDecl.Recv.List[0].Type
	var nameIndent *ast.Ident
out:
	for {
		switch recvType1 := recvType.(type) {
		case *ast.Ident:
			nameIndent = recvType1
			break out
		case *ast.StarExpr:
			recvType = recvType1.X
		case *ast.IndexExpr:
			// func (a *Clas[T])
			recvType = recvType1.X
		case *ast.IndexListExpr:
			// func (a *Clas[K ClassA,V ClassB])
			recvType = recvType1.X
		default:
			fmt.Printf("unexpected receiver type: %T in %s\n", recvType, m.GoSource.Path)
			break out
		}
	}
	// 由于代码的位置关系，这一步不一定会找到，所以自己创建了。
	receiver := m.GoSource.Pkg.FindStruct(nameIndent.Name)
	m.Receiver = receiver
	receiver.MethodManager.AddCallable(m)
	return nil
}
