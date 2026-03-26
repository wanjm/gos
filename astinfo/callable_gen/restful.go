package callable_gen

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/astinfo"
)

type RestfulGen struct {
	filters       []*FilterInfo
	filterMap     map[string]*FilterInfo
	InternalError int
	DataError     int
}

func NewRestfulGen(dataError, internalError int) *RestfulGen {
	return &RestfulGen{
		DataError:     dataError,
		InternalError: internalError,
		filterMap:     make(map[string]*FilterInfo),
	}
}

func (restful *RestfulGen) GetName() string {
	return "restful"
}

func (restful *RestfulGen) GenerateCommon(file *astinfo.GenedFile) {
	generateCommon()
}

// 定义过滤器代码生成模板
const restfulFilterTemplate = `func {{.FilterName}}(c *gin.Context) {
	err := {{.ImportName}}.{{.FunctionName}}(c, &c.Request)
	if err != nil {
		errorCode, errMessage := getErrorCode(err)
		c.JSON(errorCode, gin.H{"error": errMessage})
		c.Abort()
	}
}
`

func (restful *RestfulGen) GenFilterCode(function *astinfo.Function, file *astinfo.GenedFile) string {
	file.GetImport(astbasic.SimplePackage("github.com/gin-gonic/gin", "gin"))
	pkg := function.GoSource.Pkg
	// 生成过滤器函数名
	filterName := "filter_" + pkg.Name + "_" + function.Name
	impt := file.GetImport(&pkg.PkgBasic)

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
	tpl, err := template.New("filter").Parse(restfulFilterTemplate)
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
		if restful.filterMap == nil {
			restful.filterMap = make(map[string]*FilterInfo)
		}
		filterInfo := &FilterInfo{
			FilterName:    filterName,
			FilterRawName: function.Name,
			Func:          function,
		}
		restful.filterMap[function.Name] = filterInfo
		restful.filters = append(restful.filters, filterInfo)
		return ""
	}
}

// genRouterCode
func (restful *RestfulGen) GenRouterCode(method *astinfo.Method, file *astinfo.GenedFile) string {
	name := ""
	var sb strings.Builder
	file.AddBuilder(&sb)

	type CodeParam struct {
		HttpMethod         string
		MethodName         string
		Url                string
		FilterName         string
		RequestConstruct   string
		UrlParameterStr    string
		HeaderParameterStr string
		HasRequest         bool
		HasResponse        bool
		IsStringResponse   bool
		ResponseNilCode    string
		DataError          int
	}
	tm := &CodeParam{
		HttpMethod: method.Comment.Method,
		MethodName: method.Name,
		Url:        path.Join(method.Receiver.Comment.Url, method.Comment.Url),
		DataError:  restful.DataError,
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
		if st, ok := astinfo.GetBasicType(requestParam.Type).(*astinfo.Struct); ok {
			for _, f := range st.FieldsWithTag(astinfo.HEADER) {
				if headerName, hasHeader := f.GetHeaderName(); hasHeader {
					if rt, isRaw := astinfo.GetBasicType(f.Type).(*astinfo.RawType); isRaw && rt.RefName(nil) == "string" {
						tm.HeaderParameterStr += fmt.Sprintf("request.%s=c.GetHeader(\"%s\")\n", f.Name, headerName)
					}
				}
			}
		}
	}
	if len(method.Results) > 1 {
		tm.HasResponse = true
		rt := astinfo.GetBasicType(method.Results[0].Type)
		if raw, isRaw := rt.(*astinfo.RawType); isRaw && raw.RefName(nil) == "string" {
			tm.IsStringResponse = true
		}
	}

	//获取可能存在的url中的参数
	methodUrl := strings.Trim(method.Comment.Url, "\"")
	if strings.Contains(methodUrl, ":") {
		names := strings.SplitSeq(methodUrl, "/")
		for name := range names {
			if strings.Contains(name, ":") {
				//此处最好从名字能获取到Field，然后在调用type的parse方法，返回其对应的值；
				tm.UrlParameterStr += fmt.Sprintf("request.%s=c.Param(\"%s\")\n", astbasic.Capitalize(name[1:]), name[1:])
			}
		}
	}
	userFilters := strings.SplitSeq(method.Comment.Filter, ",")
	for filter := range userFilters {
		filter = strings.Trim(filter, "\t ")
		if filter != "" {
			if restful.filterMap != nil {
				filterInfo := restful.filterMap[filter]
				if filterInfo == nil {
					fmt.Printf("filter %s not found in file %s for %s \n", filter, method.GoSource.Path, method.Name)
				} else {
					tm.FilterName += filterInfo.FilterName + ","
				}
			} else {
				fmt.Printf("filter %s not found in file %s for %s \n", filter, method.GoSource.Path, method.Name)
			}
		}
	}
	for _, filter := range restful.filters {
		if strings.Contains(methodUrl, filter.Func.Comment.Url) {
			tm.FilterName += filter.FilterName + ","
		}
	}
	tmplText := `engine.{{.HttpMethod}} ( "{{.Url}}", {{.FilterName}} func(c *gin.Context) {
		{{ if .HasRequest }}
		request := {{.RequestConstruct}}
		{{.UrlParameterStr}}
		{{.HeaderParameterStr}}
		// 利用gin的自动绑定功能，将请求内容绑定到request对象上；兼容get,post等情况
		if err := c.ShouldBind(request); err != nil {
			c.JSON(400, gin.H{"error": "param error"})
			return
		}
		{{ end }}
		{{ if .HasResponse }}a,{{end}} err := receiver.{{.MethodName}}(c {{ if .HasRequest }},request{{ end }})
		{{.ResponseNilCode}}
		if err != nil {
			errorCode, errMessage := getErrorCode(err)
			if errorCode == 0 {
				errorCode = 500
			}
			c.JSON(errorCode, gin.H{"error": errMessage})
			return
		}
		{{ if .HasResponse }}
		{{ if .IsStringResponse }}
		c.String(200, "%s", a)
		{{ else }}
		c.JSON(200, a)
		{{ end }}
		{{ else }}
		c.Status(200)
		{{ end }}
	})
		`

	tmpl, err := template.New("restfulRouter").Parse(tmplText)
	if err != nil {
		log.Fatalf("解析模板失败: %v", err)
	}
	err = tmpl.Execute(&sb, tm)
	if err != nil {
		log.Fatalf("执行模板失败: %v", err)
	}
	return name
}
