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
	funcDecl        *ast.FuncDecl
	goSource        *Gosourse
	functionManager *FunctionManager //function指向pkg的functionManager，method指向自己recevicer(struct)的functionManager
	comment         functionComment
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

// 解析自己，并把自己添加到对应的functionManager中；
func (f *Function) Parse() error {
	if f.functionManager == nil {
		fmt.Printf("functionManager should be initialized")
	}
	parseComment(f.funcDecl.Doc, &f.comment)
	// 方法体为空
	return nil
}
