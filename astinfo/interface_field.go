package astinfo

import (
	"fmt"
	"go/ast"
)

type InterfaceFieldComment struct {
	Url string
}

func (comment *InterfaceFieldComment) dealValuePair(key, value string) {
	switch key {
	case Url:
		comment.Url = value
	default:
		fmt.Printf("unkonw key value pair => key=%s,value=%s\n", key, value)
	}
}

type InterfaceField struct {
	FunctionField
	Comment InterfaceFieldComment
	astRoot *ast.Field
}

func NewInterfaceField(field *ast.Field, goSource *Gosourse) *InterfaceField {
	return &InterfaceField{
		astRoot: field,
		FunctionField: FunctionField{
			GoSource: goSource,
		},
	}
}

// Parse 解析接口字段
func (f *InterfaceField) Parse() error {
	// 解析字段名称
	parseComment(f.astRoot.Doc, &f.Comment)
	f.Name = f.astRoot.Names[0].Name
	f.parseParameter(f.astRoot.Type.(*ast.FuncType))
	return nil
}
