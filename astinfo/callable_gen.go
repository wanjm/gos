package astinfo

type CallableGen interface {
	GetName() string //generator的name，如servlet，prpc，restful等
	GenerateCommon(file *GenedFile)
	GenFilterCode(function *Function, file *GenedFile) string
	GenRouterCode(method *Method, file *GenedFile) string
}

var callableGens []CallableGen

//	{
//		&ServletGen{},
//		&prpcGen{},
//		&ResutfulGen{},
//	}
func RegisterCallableGen(gen ...CallableGen) {
	callableGens = append(callableGens, gen...)
}
