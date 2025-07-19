package astinfo

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type ServletGen struct{}

func (servlet *ServletGen) GetName() string {
	return "servlet"
}
func (servlet *ServletGen) GenerateCommon(file *GenedFile) {
	var content strings.Builder
	Project := GlobalProject
	if Project.cfg.Generation.ResponseKey != "" {
		oneImport := file.getImport(SimplePackage(Project.cfg.Generation.ResponseMod, "xx"))
		content.WriteString("var responseKey " + oneImport.Name + "." + Project.cfg.Generation.ResponseKey)
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
	file.addBuilder(&content)
}

func (servlet *ServletGen) GenFilterCode(function *Function, file *GenedFile) string {
	file.getImport(SimplePackage("github.com/gin-gonic/gin", "gin"))
	pkg := function.goSource.pkg
	//生成这个函数，pkg.file已经生成了，所以可以直接使用
	name := "filter_" + pkg.name + "_" + function.Name
	impt := file.getImport(pkg)
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
	file.addBuilder(&sb)
	return name
}

// genRouterCode
func (servlet *ServletGen) GenRouterCode(method *Method, file *GenedFile) string {
	name := ""
	var sb strings.Builder
	file.addBuilder(&sb)
	// method.generateServletPostCall(file, &sb)
	var realParams string
	// var rawServlet = false
	//  有request请求，需要解析request，有些情况下，服务端不需要request；
	// 参数为context.Context, request *schema.Request
	paramIndex := 1
	requestParam := method.Params[paramIndex]
	if !IsPointer(requestParam.Type) {
		fmt.Print("only pointer is supported in " + strconv.Itoa(paramIndex) + " parameter(start from 0) for method " + method.Name)
		os.Exit(0)
	}
	variableCode := "request:=" + requestParam.GenVariableCode(file) + "\n"
	sb.WriteString(variableCode)
	sb.WriteString(`
		// 利用gin的自动绑定功能，将请求内容绑定到request对象上；兼容get,post等情况
		if err := engine.ShouldBind(request); err != nil {
			cJSON(engine,200, Response{
			Code: 4,
			Message: "param error",
			})
			return
		}
		`)
	realParams += ",request"

	return name
}
