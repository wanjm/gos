package astinfo

import "strings"

type CallableGen interface {
	GetName() string
	GenFilterCode(function *Function, file *GenedFile) string
	GenRouterCode(method *Method, file *GenedFile) string
}

var callableGens = []CallableGen{
	&ServletGen{},
	&prpcGen{},
	&ResutfulGen{},
}

type ServletGen struct{}

func (servlet *ServletGen) GetName() string {
	return "servlet"
}

func (servlet *ServletGen) GenFilterCode(function *Function, file *GenedFile) string {
	var content strings.Builder
	name := ""
	file.addBuilder(&content)
	return name
}

// genRouterCode
func (servlet *ServletGen) GenRouterCode(method *Method, file *GenedFile) string {
	var content strings.Builder
	name := ""
	file.addBuilder(&content)
	return name
}

type prpcGen struct{}

func (prpc *prpcGen) GetName() string {
	return "prpc"
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

type ResutfulGen struct {
}

func (resutful *ResutfulGen) GetName() string {
	return "resutful"
}

func (resutful *ResutfulGen) GenFilterCode(function *Function, file *GenedFile) string {
	var content strings.Builder
	name := ""
	file.addBuilder(&content)
	return name
}

// genRouterCode
func (resutful *ResutfulGen) GenRouterCode(method *Method, file *GenedFile) string {
	var content strings.Builder
	name := ""
	file.addBuilder(&content)
	return name
}
