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
}

const (
	POST = "POST"
	GET  = "GET"
)

type Function struct {
	Name     string //函数名
	funcDecl *ast.FuncDecl
	Comment  functionComment
	GoSource *Gosourse

	Params  []*Field // method params, 下标0是request
	Results []*Field // method results（output)	Params      []*Field // method params, 下标0是request
}

func (comment *functionComment) dealValuePair(key, value string) {
	key = strings.ToLower(key)
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
	case Title:
		comment.title = strings.Trim(value, "\"")
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
	default:
		if !comment.dealOldValuePair(key, value) {
			fmt.Printf("unknown key '%s' in function comment\n", key)
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
	return &Function{
		funcDecl: funcDecl,
		GoSource: goSource,
	}
}

// GetType() string
func (f *Function) GetType() string {
	return f.Comment.funcType
}

// 解析自己，并把自己添加到对应的functionManager中；
func (f *Function) Parse() error {
	parseComment(f.funcDecl.Doc, &f.Comment)
	f.Name = f.funcDecl.Name.Name
	f.parseParameter()
	// 方法体为空
	return nil
}

// 从ast.Field中解析出参数
func parseFields(params []*ast.Field, goSource *Gosourse) []*Field {
	var result []*Field
	for _, param := range params {
		field := Field{
			astRoot:  param,
			goSource: goSource,
		}
		field.Parse()
		if len(param.Names) != 0 {
			for _, name := range param.Names {
				field1 := field
				field1.Name = name.Name
				result = append(result, &field1)
			}
		} else {
			//没有参数名，基本不会出现
			result = append(result, &field)
		}
	}
	return result
}

// 解析参数和返回值
func (f *Function) parseParameter() bool {
	var paramType *ast.FuncType = f.funcDecl.Type
	//Params参数不可能为nil
	f.Params = parseFields(paramType.Params.List, f.GoSource)
	//Results返回值可能为nil
	if paramType.Results != nil {
		f.Results = parseFields(paramType.Results.List, f.GoSource)
	}
	return true
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
	impt := goGenerated.GetImport(f.GoSource.Pkg)
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
		Function: &Function{
			funcDecl: funcDecl,
			GoSource: goSource,
		},
		FunctionManager: &goSource.Pkg.FunctionManager,
	}
}
func (h *FunctionParserHelper) Parse() error {
	h.Function.Parse()
	h.FunctionManager.AddCallable(h.Function)
	return nil
}
