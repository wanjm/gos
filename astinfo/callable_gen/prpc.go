package callable_gen

import (
	"fmt"
	"path"
	"strings"

	"github.com/wanjm/gos/astinfo"
)

type PrpcGen struct{}

func (prpc *PrpcGen) GetName() string {
	return "prpc"
}
func (prpc *PrpcGen) GenerateCommon(file *astinfo.GenedFile) {
	file.GetImport(astinfo.SimplePackage("github.com/rs/xid", "xid"))
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
	var sb strings.Builder
	type CodeParam struct {
		HttpMethod       string
		MethodName       string
		Url              string
		FilterName       string //自带最后一个逗号
		RequestConstruct []string
		UrlParameterStr  string
		HasRequest       bool
		HasResponse      bool
		ResponseNilCode  string
		DataError        int
	}
	tm := &CodeParam{
		HttpMethod: method.Comment.Method,
		MethodName: method.Name,
		Url:        path.Join(method.Receiver.Comment.Url, method.Comment.Url),
		// DataError:  servlet.DataError,
	}
	sb.WriteString("router.POST(" + method.Comment.Url + ", func(c *gin.Context) {\n")
	var interfaceArgs string
	var realParams string
	for i := 1; i < len(method.Params); i++ {
		param := method.Params[i]
		param.GenVariableCode(file, false)
	}

	sb.WriteString(fmt.Sprintf("var request=[]interface{}{%s}\n", interfaceArgs))
	tmplText := `
	engine.{{.HttpMethod}} ( "{{.Url}}", {{.FilterName}} func(c *gin.Context) {
	{{range .RequestConstruct}}
		arg:={{.}}
	{{end}}
	{{if .HasRequest}}
	if err := c.ShouldBindJSON(&request); err != nil {
		cJSON(c,200, map[string]any{
			"o": []any{&Error{Code: 4, Message: err.Error()}},
			"c": 0,
		})
		return
	}
	var Request= c.Request;
	tid := Request.Header.Get(TraceId)
	if len(tid) ==0 {
		tid = xid.New().String()
	}
	c.Request = Request.WithContext(context.WithValue(Request.Context(), TraceIdNameInContext, tid))
	
	var code any
	if err.Code != 0 {
		code = &Error{Code: int(err.Code), Message: err.Message}
	}
	cJSON(c,200, map[string]any{
		{{if .HasResponse}}
			"o":[]any{[code,response]}
		{{else}}
			"o":[]any{code},
		{{end}}
		"c":    0,
	})
	`
	var objString string
	var objResult string
	// 返回值仅有一个是Error；
	if len(method.Results) == 2 {
		objResult = "response,"
		objString = "\"o\":[]any{[code,response},"
	} else {
		objString = "\"o\":[]any{code},"
	}
	// 返回值有两个，一个是response，一个是Error；
	// 代码暂不检查是否超过两个；
	//${objResult} err:= ${receiverPrefix}${method.Name}(c${realParams}
	sb.WriteString(fmt.Sprintf(`%s err := %s%s(c%s)
		var code any
		if err.Code != 0 {
			code = &Error{Code: int(err.Code), Message: err.Message}
		}
		cJSON(c,200, map[string]any{
			%s
			"c":    0,
		})
	`, objResult, receiverPrefix, method.Name, realParams, objString))
	sb.WriteString("})\n") //end of router.POST
	return sb.String()
	return name
}
