package callable_gen

import (
	"fmt"
	"log"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/astinfo"
)

var errorCodeGen = false

func GenBasicError(file *astinfo.GenedFile) {
	if errorCodeGen {
		return
	}
	a := `
type Error struct {
	Code    int    "json:\"code\""
	Message string "json:\"message\""
}

func (error Error) Error() string {
	return error.Message
}

type RpcResult struct {
	C int    "json:\"c\""
	O [2]any "json:\"o\""
}
	`
	var content strings.Builder
	content.WriteString(a)
	file.AddBuilder(&content)
	errorCodeGen = true
}

type PrpcGen struct{}

func (prpc *PrpcGen) GetName() string {
	return "prpc"
}
func (prpc *PrpcGen) GenerateCommon(file *astinfo.GenedFile) {
	file.GetImport(astbasic.SimplePackage("github.com/rs/xid", "xid"))
	file.GetImport(astbasic.SimplePackage("context", "context"))
	GenBasicError(file)
	generateCommon()
}

func (prpc *PrpcGen) GenFilterCode(function *astinfo.Function, file *astinfo.GenedFile) string {
	var content strings.Builder
	name := ""
	file.AddBuilder(&content)
	return name
}

// genRouterCode
func (prpc *PrpcGen) GenRouterCode(method *astinfo.Method, file *astinfo.GenedFile) string {
	var sb strings.Builder
	name := ""
	file.AddBuilder(&sb)
	type CodeParam struct {
		HttpMethod       string
		MethodName       string
		Url              string
		FilterName       string //自带最后一个逗号
		RequestConstruct []string
		UrlParameterStr  string
		HasRequest       bool
		HasResponse      bool
		// ResponseNilCode      string
		// DataError            int
		InterfaceInit        string
		ParamString          string
		TraceIdNameInContext string
	}
	tm := &CodeParam{
		HttpMethod: method.Comment.Method,
		MethodName: method.Name,
		Url:        path.Join(method.Receiver.Comment.Url, method.Comment.Url),
		// DataError:  servlet.DataError,
	}
	var args []string
	var params []string
	for i := 1; i < len(method.Params); i++ {
		param := method.Params[i]
		argname := "arg" + strconv.Itoa(i)
		tm.RequestConstruct = append(tm.RequestConstruct, argname+":="+param.GenVariableCode(file, false))
		args = append(args, "&"+argname)
		params = append(params, argname)
	}
	tm.ParamString = strings.Join(params, ",")
	tm.InterfaceInit = strings.Join(args, ",")
	tm.HasRequest = len(method.Params) > 0
	tm.HasResponse = len(method.Results) > 1
	generationCfg := &astinfo.GlobalProject.Cfg.Generation
	if generationCfg.TraceKey != "" {
		// prpc的发送请求是，会向http头添加traceId，需要使用该变量
		oneImport := file.GetImport(astbasic.SimplePackage(generationCfg.TraceKeyMod, "xx"))
		tm.TraceIdNameInContext = fmt.Sprintf("%s.%s{}", oneImport.Name, generationCfg.TraceKey)
	} else {
		tm.TraceIdNameInContext = `"badTraceIdName plase config in Generation TraceKeyMod"`
	}
	tmplText := `
	engine.{{.HttpMethod}} ( "{{.Url}}", {{.FilterName}} func(c *gin.Context) {
	{{if .HasRequest}}
		{{range .RequestConstruct}}
			{{.}} 
		{{end}}
		request:=[]any{ {{.InterfaceInit}} }
	if err := c.ShouldBindJSON(&request); err != nil {
		cJSON(c,200, map[string]any{
			"o": []any{&Error{Code: 4, Message: err.Error()}},
			"c": 0,
		})
		return
	}
	{{end}}
	var Request= c.Request;
	tid := Request.Header.Get(TraceId)
	if len(tid) ==0 {
		tid = xid.New().String()
	}
	c.Request = Request.WithContext(context.WithValue(Request.Context(), {{.TraceIdNameInContext}}, tid))
	{{ if .HasResponse }}response,{{end}} err := receiver.{{.MethodName}}(c {{ if .HasRequest }},{{.ParamString}}{{ end }})
	var code any
	errorCode,errMessage:=getErrorCode(err)
			if errorCode != 0 {
			code = &Error{Code: errorCode, Message: errMessage}
		}
	cJSON(c,200, map[string]any{
		{{if .HasResponse}}
		"o": []any{code, response},
		{{else}}
		"o": []any{code},	
		{{end}}
		"c":    0,
	})
})
	`
	tmpl, err := template.New("personInfo").Parse(tmplText)
	if err != nil {
		log.Fatalf("解析rpc模板失败: %v", err)
	}
	err = tmpl.Execute(&sb, tm)
	if err != nil {
		log.Fatalf("执行rpc模板失败: %v", err)
	}
	return name
}
