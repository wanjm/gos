package astinfo

import (
	"fmt"
	"go/ast"
	"strings"
)

type InterfaceFieldComment struct {
	Url    string
	Method string // HTTP method for restrpc; default POST
}

func (comment *InterfaceFieldComment) dealValuePair(key, value string) {
	switch key {
	case Url:
		comment.Url = value
	case ConstMethod:
		comment.Method = strings.ToUpper(strings.Trim(value, "\""))
		if comment.Method != "" {
			if _, ok := methodMap[comment.Method]; !ok {
				fmt.Printf("method '%s' is not supported in interface method comment\n", comment.Method)
			}
		}
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
