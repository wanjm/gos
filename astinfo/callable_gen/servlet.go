package callable_gen

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/wanjm/gos/astinfo"
	"github.com/wanjm/gos/tool"
)

type FilterInfo struct {
	FilterName    string
	FilterRawName string
	Func          *astinfo.Function
}
type ServletGen struct {
	filters       []*FilterInfo
	filterMap     map[string]*FilterInfo
	InternalError int
	DataError     int
}

func NewServletGen(dataError, internalError int) *ServletGen {
	servlet := &ServletGen{
		DataError:     dataError,
		InternalError: internalError,
		filterMap:     make(map[string]*FilterInfo),
	}
	return servlet
}
func (servlet *ServletGen) GetName() string {
	return "servlet"
}

func (servlet *ServletGen) GenerateCommon(file *astinfo.GenedFile) {
	generateCommon()
}

// 定义过滤器代码生成模板
const filterTemplate = `func {{.FilterName}}(c *gin.Context) {
	err := {{.ImportName}}.{{.FunctionName}}(c, &c.Request)
	errorCode,errMessage:=getErrorCode(err)
	if errorCode != 0 {
		cJSON(c, 200, Response{
			Code:    errorCode,
			Message: errMessage,
		})
		c.Abort()
	}
}
`

func (servlet *ServletGen) GenFilterCode(function *astinfo.Function, file *astinfo.GenedFile) string {
	file.GetImport(astinfo.SimplePackage("github.com/gin-gonic/gin", "gin"))
	pkg := function.GoSource.Pkg
	// 生成过滤器函数名
	filterName := "filter_" + pkg.Name + "_" + function.Name
	impt := file.GetImport(pkg)

	// 准备模板数据
	data := struct {
		FilterName   string
		ImportName   string
		FunctionName string
	}{}
	data.FilterName = filterName
	data.ImportName = impt.Name
	data.FunctionName = function.Name

	// 解析并执行模板
	var sb strings.Builder
	tpl, err := template.New("filter").Parse(filterTemplate)
	if err != nil {
		panic(err)
	}
	if err := tpl.Execute(&sb, data); err != nil {
		panic(err)
	}

	file.AddBuilder(&sb)

	// 处理URL注释逻辑
	if function.Comment.Url == "" || function.Comment.Url == "\"\"" {
		return filterName
	} else {
		filterInfo := &FilterInfo{
			FilterName:    filterName,
			FilterRawName: function.Name,
			Func:          function,
		}
		servlet.filterMap[function.Name] = filterInfo
		servlet.filters = append(servlet.filters, filterInfo)
		return ""
	}
}

// genRouterCode
func (servlet *ServletGen) GenRouterCode(method *astinfo.Method, file *astinfo.GenedFile) string {
	name := ""
	var sb strings.Builder
	file.AddBuilder(&sb)
	// method.generateServletPostCall(file, &sb)
	// var realParams string
	// var rawServlet = false
	//  有request请求，需要解析request，有些情况下，服务端不需要request；
	// 参数为context.Context, request *schema.Request
	type CodeParam struct {
		HttpMethod       string
		MethodName       string
		Url              string
		FilterName       string //自带最后一个逗号
		RequestConstruct string
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
		DataError:  servlet.DataError,
	}
	if len(method.Params) > 1 {
		paramIndex := 1
		requestParam := method.Params[paramIndex]
		if !astinfo.IsPointer(requestParam.Type) {
			fmt.Printf("only pointer is supported in %s of file %s \n", method.Name, method.GoSource.Path)
			os.Exit(0)
		}
		tm.HasRequest = true
		tm.RequestConstruct = requestParam.GenVariableCode(file, false)
	}
	if len(method.Results) > 1 {
		tm.HasResponse = true
		// tm.ResponseNilCode = method.Results[0].GenNilCode(file)
	}

	//获取可能存在的url中的参数
	methodUrl := strings.Trim(method.Comment.Url, "\"")
	if strings.Contains(methodUrl, ":") {
		names := strings.Split(methodUrl, "/")
		for _, name := range names {
			if strings.Contains(name, ":") {
				//此处最好从名字能获取到Field，然后在调用type的parse方法，返回其对应的值；
				tm.UrlParameterStr += fmt.Sprintf("request.%s=c.Param(\"%s\")\n", tool.Capitalize(name[1:]), name[1:])
			}
		}
	}
	userFilters := strings.Split(method.Comment.Filter, ",")
	for _, filter := range userFilters {
		filter = strings.Trim(filter, "\t ")
		if filter != "" {
			filterInfo := servlet.filterMap[filter]
			if filterInfo == nil {
				fmt.Printf("filter %s not found in file %s for %s \n", filter, method.GoSource.Path, method.Name)
			} else {
				tm.FilterName += filterInfo.FilterName + ","
			}
		}
	}
	for _, filter := range servlet.filters {
		if strings.Contains(methodUrl, filter.Func.Comment.Url) {
			tm.FilterName += filter.FilterName + ","
		}
	}
	tmplText := `engine.{{.HttpMethod}} ( "{{.Url}}", {{.FilterName}} func(c *gin.Context) {
		{{ if .HasRequest }}
		request := {{.RequestConstruct}}
		{{.UrlParameterStr}}	
		// 利用gin的自动绑定功能，将请求内容绑定到request对象上；兼容get,post等情况
		if err := c.ShouldBind(request); err != nil {
			cJSON(c, 200, Response{
				Code:    {{.DataError}},
				Message: "param error",
			})
			return
		}
		{{ end }}
		{{ if .HasResponse }}a,{{end}} err := receiver.{{.MethodName}}(c {{ if .HasRequest }},request{{ end }})
		{{.ResponseNilCode}}
		var code = 200
		errorCode,errMessage:=getErrorCode(err)
		var extraInfo any
		if exta, ok := err.(ExtraInfo); ok {
			extraInfo = exta.GetExtraInfo()
		}
		cJSON(c, code, Response{
			{{ if .HasResponse }}Object:  a,{{ end }}
			Code:    errorCode,
			ExtraInfo: extraInfo,
			Message: errMessage,
		})
	})
		`

	tmpl, err := template.New("personInfo").Parse(tmplText)
	if err != nil {
		log.Fatalf("解析模板失败: %v", err)
	}
	err = tmpl.Execute(&sb, tm)
	if err != nil {
		log.Fatalf("执行模板失败: %v", err)
	}
	return name
}
