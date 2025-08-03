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
	// 如果function类型为空，则不继续解析
	if m.Comment.funcType != "" {
		return m.parseReceiver()
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
	// 虽然现在先解析类型，在解析函数，但是只能再一个文件内容保持这个顺序，如果定义在多个文件，还是不能保证结构体肯定存在；
	receiver := m.GoSource.Pkg.GetTyper(nameIndent.Name).(*Struct)
	m.Receiver = receiver
	receiver.MethodManager.AddCallable(m)
	return nil
}
