package callable_gen

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/wan_jm/servlet/astinfo"
)

type FilterInfo struct {
	FilterName    string
	FilterRawName string
	Func          *astinfo.Function
}
type ServletGen struct {
	filters []*FilterInfo
}

func (servlet *ServletGen) GetName() string {
	return "servlet"
}

var commongened bool

func (servlet *ServletGen) GenerateCommon(file *astinfo.GenedFile) {
	if commongened {
		return
	}
	commongened = true
	var content strings.Builder
	Project := astinfo.GlobalProject
	if Project.Cfg.Generation.ResponseKey != "" {
		oneImport := file.GetImport(astinfo.SimplePackage(Project.Cfg.Generation.ResponseMod, "xx"))
		content.WriteString("var responseKey " + oneImport.Name + "." + Project.Cfg.Generation.ResponseKey)
		content.WriteString(`
	type JsonString struct {
		context context.Context
		data    any
	}

		func (r JsonString) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	jsonBytes, err := json.Marshal(r.data)
	if err != nil {
		return err
	}
	v := r.context.Value(responseKey)
	if v != nil {
		*(v.(*string)) = string(jsonBytes)
	}
	_, err = w.Write(jsonBytes)
	return err
}

// WriteContentType (JSON) writes JSON ContentType.
func (r JsonString) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"application/json; charset=utf-8"}
	}
}

func cJSON(c *gin.Context, code int,response any) {
	c.Render(code, JsonString{
		context: c,
		data:    response,
	})
}
	`)

	} else {
		content.WriteString("func cJSON(c *gin.Context, code int, response any) {\n")
		content.WriteString("c.JSON(code, response)\n")
		content.WriteString("}\n")
	}
	file.AddBuilder(&content)
}

func (servlet *ServletGen) GenFilterCode(function *astinfo.Function, file *astinfo.GenedFile) string {
	file.GetImport(astinfo.SimplePackage("github.com/gin-gonic/gin", "gin"))
	pkg := function.GoSource.Pkg
	//生成这个函数，pkg.file已经生成了，所以可以直接使用
	name := "filter_" + pkg.Name + "_" + function.Name
	impt := file.GetImport(pkg)
	var sb = strings.Builder{}
	sb.WriteString("// filter_${pkg.file.name}_${filter.function.Name}\n")
	sb.WriteString("func ")
	sb.WriteString(name)
	sb.WriteString("(c *gin.Context) {\nres:=")
	sb.WriteString(impt.Name + "." + function.Name)
	sb.WriteString(`(c,&c.Request)
	if(res.Code!=0){
			cJSON(c,200, 
			Response{
				Code:int(res.Code),
				Message: res.Message,
			})
			c.Abort()
		}
	}
	`)
	file.AddBuilder(&sb)
	if (function.Comment.Url) == "" {
		return name
	} else {
		servlet.filters = append(servlet.filters, &FilterInfo{
			FilterName:    name,
			FilterRawName: function.Name,
			Func:          function,
		})
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
	}
	tm := &CodeParam{
		HttpMethod: method.Comment.Method,
		MethodName: method.Name,
		Url:        method.Comment.Url,
	}
	if len(method.Params) > 1 {
		paramIndex := 1
		requestParam := method.Params[paramIndex]
		if !astinfo.IsPointer(requestParam.Type) {
			fmt.Print("only pointer is supported in " + strconv.Itoa(paramIndex) + " parameter(start from 0) for method " + method.Name)
			os.Exit(0)
		}
		tm.HasRequest = true
		tm.RequestConstruct = requestParam.GenVariableCode(file, false)
	}
	if len(method.Results) > 1 {
		tm.HasResponse = true
	}

	//获取可能存在的url中的参数
	methodUrl := strings.Trim(method.Comment.Url, "\"")
	if strings.Contains(methodUrl, ":") {
		names := strings.Split(methodUrl, "/")
		for _, name := range names {
			if strings.Contains(name, ":") {
				//此处最好从名字能获取到Field，然后在调用type的parse方法，返回其对应的值；
				tm.UrlParameterStr += fmt.Sprintf("request.%s=c.Param(\"%s\")\n", astinfo.Capitalize(name[1:]), name[1:])
			}
		}
	}
	for _, filter := range servlet.filters {
		if filter.FilterRawName == method.Name {
			tm.FilterName += filter.FilterName + ","
		} else if strings.Contains(methodUrl, filter.Func.Comment.Url) {
			tm.FilterName += filter.FilterName + ","
		}
	}
	tmplText := `engine.{{.HttpMethod}} ( {{.Url}}, {{.FilterName}} func(c *gin.Context) {
		{{ if .HasRequest }}
		request := {{.RequestConstruct}}
		{{.UrlParameterStr}}	
		// 利用gin的自动绑定功能，将请求内容绑定到request对象上；兼容get,post等情况
		if err := c.ShouldBind(request); err != nil {
			cJSON(c, 200, Response{
				Code:    4,
				Message: "param error",
			})
			return
		}
		{{ end }}
		{{ if .HasResponse }}response,{{end}} err := receiver.{{.MethodName}}(c {{ if .HasRequest }},request{{ end }})
		var code = 200

		cJSON(c, code, Response{
			{{ if .HasResponse }}Object:  response,{{ end }}
			Code:    int(err.Code),
			Message: err.Message,
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
