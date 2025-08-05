package callable_gen

import (
	"strings"

	"github.com/wanjm/go_servlet/astinfo"
)

type ResutfulGen struct {
}

func (restful *ResutfulGen) GetName() string {
	return "restful"
}
func (restful *ResutfulGen) GenerateCommon(file *astinfo.GenedFile) {
}

func (restful *ResutfulGen) GenFilterCode(function *astinfo.Function, file *astinfo.GenedFile) string {
	var content strings.Builder
	name := ""
	file.AddBuilder(&content)
	return name
}

// genRouterCode
func (restful *ResutfulGen) GenRouterCode(method *astinfo.Method, file *astinfo.GenedFile) string {
	var content strings.Builder
	name := ""
	file.AddBuilder(&content)
	return name
}
