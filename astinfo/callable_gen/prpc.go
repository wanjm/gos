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
	Code    int             "json:\"code\""
	Message string          "json:\"message\""
	Obj     json.RawMessage "json:\"obj\""
}

func (error Error) Error() string {
	return error.Message
}

func (error Error) GetErrorCode() int {
	return error.Code
}

func (error Error) GetObj() json.RawMessage {
	return error.Obj
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

type PrpcGen struct {
	filters   []*FilterInfo
	filterMap map[string]*FilterInfo
}

func (prpc *PrpcGen) GetName() string {
	return "prpc"
}
func (prpc *PrpcGen) GenerateCommon(file *astinfo.GenedFile) {
	file.GetImport(astbasic.SimplePackage("github.com/rs/xid", "xid"))
	file.GetImport(astbasic.SimplePackage("context", "context"))
	GenBasicError(file)
	generateCommon()
}

const prpcFilterTemplate = `func {{.FilterName}}(c *gin.Context) {
	err := {{.ImportName}}.{{.FunctionName}}(c, &c.Request)
	errorCode, errMessage := getErrorCode(err)
	if errorCode != 0 {
		cJSON(c, 200, map[string]any{
			"o": []any{&Error{Code: errorCode, Message: errMessage}},
			"c": 0,
		})
		c.Abort()
	}
}
`

// GenFilterCode writes the gin wrapper for a prpc-group filter.
// Empty-url filters become server middleware; others attach to matching routes.
func (prpc *PrpcGen) GenFilterCode(function *astinfo.Function, file *astinfo.GenedFile) string {
	file.GetImport(astbasic.SimplePackage("github.com/gin-gonic/gin", "gin"))
	pkg := function.GoSource.Pkg
	filterName := "filter_" + pkg.Name + "_" + function.Name
	impt := file.GetImport(&pkg.PkgBasic)

	data := struct {
		FilterName   string
		ImportName   string
		FunctionName string
	}{
		FilterName:   filterName,
		ImportName:   impt.Name,
		FunctionName: function.Name,
	}

	var sb strings.Builder
	tpl, err := template.New("prpc_filter").Parse(prpcFilterTemplate)
	if err != nil {
		panic(err)
	}
	if err := tpl.Execute(&sb, data); err != nil {
		panic(err)
	}
	file.AddBuilder(&sb)

	if function.Comment.Url == "" || function.Comment.Url == "\"\"" {
		return filterName
	}
	if prpc.filterMap == nil {
		prpc.filterMap = make(map[string]*FilterInfo)
	}
	filterInfo := &FilterInfo{
		FilterName:    filterName,
		FilterRawName: function.Name,
		Func:          function,
	}
	prpc.filterMap[function.Name] = filterInfo
	prpc.filters = append(prpc.filters, filterInfo)
	return ""
}

// appendRouteFilterNames collects explicit and URL-matched filter names for one route.
// It returns a comma-terminated list so the gin handler can be appended after them.
func (prpc *PrpcGen) appendRouteFilterNames(method *astinfo.Method) string {
	var names string
	methodUrl := strings.Trim(method.Comment.Url, "\"")
	for _, filter := range method.Comment.Filters {
		if prpc.filterMap == nil {
			fmt.Printf("filter %s not found in file %s for %s \n", filter, method.GoSource.Path, method.Name)
			continue
		}
		filterInfo := prpc.filterMap[filter]
		if filterInfo == nil {
			fmt.Printf("filter %s not found in file %s for %s \n", filter, method.GoSource.Path, method.Name)
			continue
		}
		names += filterInfo.FilterName + ","
	}
	for _, filter := range prpc.filters {
		if strings.Contains(methodUrl, filter.Func.Comment.Url) {
			names += filter.FilterName + ","
		}
	}
	return names
}

// GenRouterCode writes one prpc gin route, including per-route filters.
// It keeps the built-in TraceId only when a filter has not already set one.
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
		InterfaceInit string
		ParamString   string
	}
	tm := &CodeParam{
		HttpMethod: method.Comment.Method,
		MethodName: method.Name,
		Url:        path.Join(method.Receiver.Comment.Url, method.Comment.Url),
		FilterName: prpc.appendRouteFilterNames(method),
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
	if c.Value(TraceIdNameInContext) == nil {
		var Request= c.Request;
		tid := Request.Header.Get(TraceId)
		if len(tid) ==0 {
			tid = xid.New().String()
		}
		c.Request = Request.WithContext(context.WithValue(Request.Context(), TraceIdNameInContext, tid))
	}
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
	tmpl, err := template.New("prpc_router").Parse(tmplText)
	if err != nil {
		log.Fatalf("解析rpc模板失败: %v", err)
	}
	err = tmpl.Execute(&sb, tm)
	if err != nil {
		log.Fatalf("执行rpc模板失败: %v", err)
	}
	return name
}
