package astinfo

import (
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
	// var sb strings.Builder
	// // method.generateServletPostCall(file, &sb)
	// var realParams string
	// var rawServlet = false
	// //  有request请求，需要解析request，有些情况下，服务端不需要request；
	// if len(method.Params) >= 2 {
	// 	var variableCode string
	// 	methodUrl := strings.Trim(method.comment.Url, "\"")
	// 	paramIndex := 1
	// 	if strings.Contains(methodUrl, ":") {
	// 		requestParam := method.Params[paramIndex]
	// 		realParams += "," + requestParam.name
	// 		variableCode = requestParam.name + ":=" + requestParam.generateCode(receiverPrefix, file) + "\n"
	// 		sb.WriteString(variableCode)
	// 		names := strings.Split(methodUrl, "/")
	// 		for _, name := range names {
	// 			if strings.Contains(name, ":") {
	// 				sb.WriteString(fmt.Sprintf("%s.%s=c.Param(\"%s\")\n", requestParam.name, name[1:], name[1:]))
	// 			}
	// 		}
	// 		paramIndex = 2
	// 	}
	// 	requestParam := method.Params[paramIndex]
	// 	if requestParam.pkg.modPath == "net/http" {
	// 		// 此处是临时添加的解决第三方回调的问题；
	// 		// 具体如何识别，1. 参数直接使用http.Request；
	// 		// 2. 返回值改为basic.HTTPError.(Code表示http头的code，response就是完整的报文体) 这个更好。这样第一条就可以是根据需要传入
	// 		realParams += ",c.Request"
	// 		rawServlet = true
	// 	} else {
	// 		if !requestParam.isPointer {
	// 			fmt.Print("only pointer is supported in " + strconv.Itoa(paramIndex) + " parameter(start from 0) for method " + method.Name)
	// 			os.Exit(0)
	// 		}
	// 		variableCode = "request:=" + requestParam.generateCode(receiverPrefix, file) + "\n"
	// 		sb.WriteString(variableCode)
	// 		sb.WriteString(`
	// 	// 利用gin的自动绑定功能，将请求内容绑定到request对象上；兼容get,post等情况
	// 	if err := c.ShouldBind(request); err != nil {
	// 		cJSON(c,200, Response{
	// 		Code: 4,
	// 		Message: "param error",
	// 		})
	// 		return
	// 	}
	// 	`)
	// 		realParams += ",request"
	// 	}
	// }
	// var objString string
	// var objName string
	// var objResult string
	// // 返回值仅有一个是Error；
	// if len(method.Results) == 2 {
	// 	objName = "response"
	// 	objResult = objName + ","
	// 	objString = "Object:response,"
	// }
	// // sb.WriteString(method.genTraceId(file))
	// // 返回值有两个，一个是response，一个是Error；
	// // 代码暂不检查是否超过两个；
	// sb.WriteString(fmt.Sprintf("%s err := %s%s(c%s)\n", objResult, receiverPrefix, method.Name, realParams))
	// //realParams后续考虑使用strings.Join()来处理；潜力基本挖光了
	// //此处后续考虑解析参数格式，然后添加正确的写入顺序
	// if postAction, ok := method.funcManager.postAction[method.Name]; ok {
	// 	sb.WriteString(fmt.Sprintf("%s%s(c%s,%serr)\n", receiverPrefix, postAction.Name, realParams, objResult))
	// }
	// if rawServlet {
	// 	sb.WriteString("cJSON(c,int(err.Code), response)")
	// } else {
	// 	sb.WriteString("var code=200;\n")
	// 	if method.comment.method == "GET" {
	// 		sb.WriteString(`
	// 	if err.Code==500 {
	// 		// 临时兼容health check;
	// 		code=500
	// 	}
	// `)
	// 	}
	// 	if len(objName) > 0 {
	// 		var a = *method.Results[0]
	// 		a.name = objName
	// 		if a.isPointer {
	// 			panic("response should not be pointer in " + receiverPrefix + method.Name)
	// 		}
	// 		a.genCheckArrayNil("", file, &sb)
	// 	}
	// 	sb.WriteString(fmt.Sprintf(`
	// 	cJSON(c,code, Response{
	// 		%s
	// 		Code:   int(err.Code),
	// 		Message: err.Message,
	// 	})
	// `, objString))
	// }
	// sb.WriteString("})\n")

	return name
}
