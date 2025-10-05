package astinfo

import (
	"fmt"
	"go/ast"
	"strings"
)

type functionComment struct {
	Url          string // url
	title        string // 函数描述，供swagger使用
	Method       string // http方法，GET,POST，默认是POST
	isDeprecated bool
	funcType     string //函数类型，filter，servlet，websocket，prpc，initiator,creator
	security     []string
	groupName    string
	Filter       string
	owner        *Function
}

const (
	POST    = "POST"
	GET     = "GET"
	DELETE  = "DELETE"
	PUT     = "PUT"
	PATCH   = "PATCH"
	OPTIONS = "OPTIONS"
	HEAD    = "HEAD"
)

var methodMap = map[string]string{
	POST:    POST,
	GET:     GET,
	DELETE:  DELETE,
	PUT:     PUT,
	PATCH:   PATCH,
	OPTIONS: OPTIONS,
	HEAD:    HEAD,
}

type Function struct {
	funcDecl *ast.FuncDecl
	Comment  functionComment
	FunctionField
}

func (comment *functionComment) dealValuePair(key, value string) {
	key = strings.ToLower(key)
	value = strings.Trim(value, "\"")
	switch key {
	case Url:
		comment.Url = value
		if len(comment.funcType) == 0 {
			//默认是servlet
			comment.funcType = Servlet
			if len(comment.Method) == 0 {
				comment.Method = POST
			}
		}
	case Security:
		comment.security = strings.Split(value, ",")
	case ConstMethod:
		comment.Method = strings.ToUpper(value)
		if _, ok := methodMap[comment.Method]; !ok {
			fmt.Printf("method '%s' is not supported in function comment %s in %s\n", comment.Method, comment.owner.Name, comment.owner.GoSource.Path)
		}
	case Title:
		comment.title = value
	case Type:
		comment.funcType = value
		if value == Websocket {
			comment.Method = GET
		}
	case Group:
		comment.groupName = value
	case FilterConst:
		comment.groupName = value
		comment.funcType = FilterConst
	case UserFilter:
		comment.Filter = value
	default:
		if !comment.dealOldValuePair(key, value) {
			fmt.Printf("unknown key '%s' in function comment %s in %s\n", key, comment.owner.Name, comment.owner.GoSource.Path)
		}
	}
}
func (comment *functionComment) dealOldValuePair(key, value string) bool {
	switch key {
	case Creator:
		comment.funcType = ConstMethod
	case UrlFilter:
		comment.Url = value
		comment.funcType = FilterConst
	case FilterConst:
		if len(value) == 0 {
			value = Servlet
		}
		comment.funcType = FilterConst
	case Servlet:
		comment.funcType = Servlet
	case Prpc:
		comment.funcType = Prpc
	case Initiator:
		comment.funcType = Initiator
	case Websocket:
		comment.Method = GET
		comment.funcType = Websocket
	default:
		return false
	}
	return true
}

// create
func NewFunction(funcDecl *ast.FuncDecl, goSource *Gosourse) *Function {
	fun := &Function{
		funcDecl: funcDecl,
		FunctionField: FunctionField{
			GoSource: goSource,
		},
	}
	fun.Comment.owner = fun
	return fun
}

// GetType() string
func (f *Function) GetType() string {
	return f.Comment.funcType
}

// 解析自己，并把自己添加到对应的functionManager中；
func (f *Function) Parse() error {
	parseComment(f.funcDecl.Doc, &f.Comment)
	f.Name = f.funcDecl.Name.Name
	//没有类型的函数，不解析；
	if f.Comment.funcType != "" {
		f.parseParameter(f.funcDecl.Type)
	}
	return nil
}

// GenerateDependcyCode 生成依赖代码
func (f *Function) GenerateDependcyCode(goGenerated *GenedFile) string {
	return f.GenerateCallCode(goGenerated)
}

// generateCallCode 生成调用代码
// 生成pkg.functionName(var1,var2);
// 同步生成import语句；、
//
// 调用场景
// 1. initvarialbe中调用initor函数；
// 2. 在生成变量时，可能需要使用crator来生成；
func (f *Function) GenerateCallCode(goGenerated *GenedFile) string {
	var call strings.Builder
	impt := goGenerated.GetImport(&f.GoSource.Pkg.PkgBasic)
	call.WriteString(impt.Name + "." + f.Name)
	call.WriteString("(")
	for i, param := range f.Params {
		if i != 0 {
			call.WriteString(", ")
		}
		variable := Variable{
			Type: param.Type,
			Name: param.Name,
		}
		// 目前仅遇到function initiator调用的情况，所以直接找名字；
		variableName := variable.Generate(goGenerated)
		call.WriteString(variableName)
	}
	call.WriteString(")")
	return call.String()
}

type FunctionParserHelper struct {
	*Function
	*FunctionManager
}

func NewFunctionParserHelper(funcDecl *ast.FuncDecl, goSource *Gosourse) *FunctionParserHelper {
	return &FunctionParserHelper{
		Function:        NewFunction(funcDecl, goSource),
		FunctionManager: &goSource.Pkg.FunctionManager,
	}
}
func (h *FunctionParserHelper) Parse() error {
	h.Function.Parse()
	h.FunctionManager.AddCallable(h.Function)
	return nil
}
