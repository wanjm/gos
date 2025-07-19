package astinfo

import "strings"

type prpcGen struct{}

func (prpc *prpcGen) GetName() string {
	return "prpc"
}
func (prpc *prpcGen) GenerateCommon(file *GenedFile) {
}

func (prpc *prpcGen) GenFilterCode(function *Function, file *GenedFile) string {
	var content strings.Builder
	name := ""
	file.addBuilder(&content)
	return name
}

// genRouterCode
func (prpc *prpcGen) GenRouterCode(method *Method, file *GenedFile) string {
	var content strings.Builder
	name := ""
	file.addBuilder(&content)
	return name
}
