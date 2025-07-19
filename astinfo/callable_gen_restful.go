package astinfo

import "strings"

type ResutfulGen struct {
}

func (restful *ResutfulGen) GetName() string {
	return "restful"
}
func (restful *ResutfulGen) GenerateCommon(file *GenedFile) {
}

func (restful *ResutfulGen) GenFilterCode(function *Function, file *GenedFile) string {
	var content strings.Builder
	name := ""
	file.addBuilder(&content)
	return name
}

// genRouterCode
func (restful *ResutfulGen) GenRouterCode(method *Method, file *GenedFile) string {
	var content strings.Builder
	name := ""
	file.addBuilder(&content)
	return name
}
