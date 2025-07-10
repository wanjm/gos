package astinfo

import (
	"fmt"
	"go/ast"
	"strings"
)

type functionComment struct {
	Url          string // url
	title        string // 函数描述，供swagger使用
	method       string // http方法，GET,POST，默认是POST
	isDeprecated bool
	funcType     string //函数类型，filter，servlet，websocket，prpc，initiator,creator
	security     []string
}

const (
	POST = "POST"
	GET  = "GET"
)

type Function struct {
	Name     string //函数名
	funcDecl *ast.FuncDecl
	goSource *Gosourse
	comment  functionComment

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
			if len(comment.method) == 0 {
				comment.method = POST
			}
		}
	case Security:
		comment.security = strings.Split(value, ",")
	case ConstMethod:
		comment.method = strings.ToUpper(value)
	case Title:
		comment.title = strings.Trim(value, "\"")
	case Type:
		comment.funcType = value
		if value == Websocket {
			comment.method = GET
		}
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
		comment.method = GET
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
		goSource: goSource,
	}
}

// GetType() string
func (f *Function) GetType() string {
	return f.comment.funcType
}

// 解析自己，并把自己添加到对应的functionManager中；
func (f *Function) Parse() error {
	parseComment(f.funcDecl.Doc, &f.comment)
	f.parseParameter()
	// 方法体为空
	return nil
}

// 解析参数和返回值
func (f *Function) parseParameter() bool {
	var paramType *ast.FuncType = f.funcDecl.Type
	for _, param := range paramType.Params.List {
		// field := Field{
		// 	ownerInfo: "function Name is " + method.Name,
		// }
		// field.parseType(param.Type, method.goSource)
		//此处可能多个参数 a,b string的格式暂时仅处理一个；
		for _, name := range param.Names {
			_ = name
			// nfield := field
			// nfield.name = name.Name
			// method.Params = append(method.Params, &nfield)
			// break
		}
	}
	if paramType.Results != nil {
		for _, result := range paramType.Results.List {

			// field := Field{
			// 	ownerInfo: "function Name is " + method.Name,
			// }
			// field.parseType(result.Type, method.goFile)

			if len(result.Names) != 0 {
				// field.name = result.Names[0].Name
			}
			// method.Results = append(method.Results, &field)
		}
	}
	return true
}

// generateCallCode 生成调用代码
func (f *Function) GenerateCallCode(goGenerated *GenedFile) string {
	var call strings.Builder
	impt := goGenerated.getImport(f.goSource.pkg)
	call.WriteString(impt.Name + "." + f.Name)
	call.WriteString("(")
	for i, param := range f.Params {
		if i != 0 {
			call.WriteString(", ")
		}
		call.WriteString(param.Name)
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
			goSource: goSource,
		},
		FunctionManager: &goSource.pkg.FunctionManager,
	}
}
func (h *FunctionParserHelper) Parse() error {
	h.Function.Parse()
	h.FunctionManager.AddCallable(h.Function)
	return nil
}
