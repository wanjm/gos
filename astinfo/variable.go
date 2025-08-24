package astinfo

import (
	"fmt"
	"strings"
)

// Field是源码中定义的变量；
// Variable时是生成的代码中定义的用来使用的变量；
type Variable struct {
	Type Typer
	Name string
	Wire bool
}

// 当需要一个变量值时如下几个场景；
// 该变量类型在全局函数存在，则从全局变量获取，直接返回变量名即可
// reciver.function creator!=nil, receiverPrex!=""
// schema.struct
// schema.function  creator!=nil, receiverPrefix==""
// 返回值无\n
func (v *Variable) Generate(goGenerated *GenedFile) string {
	var variableCode = v.genFromGlobal(goGenerated)
	if variableCode != "" {
		return variableCode
	}

	variableCode = v.Type.GenConstructCode(goGenerated, v.Wire)
	return variableCode
}

// genFromGlobal
func (v *Variable) genFromGlobal(_ *GenedFile) string {
	var variableCode string
	variableNode := GlobalProject.GetVariableNode(v.Type, v.Name)
	if variableNode != nil {
		variableCode = variableNode.returnVariableName
		returnField := variableNode.getReturnField()
		var returnDepth = PointerDepth(returnField.Type)
		var targetDepth = PointerDepth(v.Type)
		var delta = returnDepth - targetDepth
		if delta < 0 {
			if delta != -1 {
				fmt.Printf("")
			}
			variableCode = "&" + variableCode
		} else {
			variableCode = strings.Repeat("*", delta) + variableCode
		}
	}
	return variableCode
}
