package callable_gen

import (
	"strings"

	"github.com/wanjm/go_servlet/astinfo"
)

type PrpcGen struct{}

func (prpc *PrpcGen) GetName() string {
	return "prpc"
}
func (prpc *PrpcGen) GenerateCommon(file *astinfo.GenedFile) {
}

func (prpc *PrpcGen) GenFilterCode(function *astinfo.Function, file *astinfo.GenedFile) string {
	var content strings.Builder
	name := ""
	file.AddBuilder(&content)
	return name
}

// genRouterCode
func (prpc *PrpcGen) GenRouterCode(method *astinfo.Method, file *astinfo.GenedFile) string {
	var content strings.Builder
	name := ""
	file.AddBuilder(&content)
	return name
}
