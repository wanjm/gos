package callable_gen

import (
	"strings"
	"text/template"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/astinfo"
)

var commongened bool

// 定义代码生成模板
const cJsonTemplate = `
func dealErrorResult(err error, c *gin.Context, code int, errorCode int, errMessage string) {
	var extraInfo any
	if exta, ok := err.(ExtraInfo); ok {
		extraInfo = exta.GetExtraInfo()
	}
	cJSON(c, code, Response{
		Code:      errorCode,
		ExtraInfo: extraInfo,
		Message:   errMessage,
	})
}
{{if .HasResponseKey}}
var responseKey {{.ImportName}}.{{.ResponseKey}}
{{end}}

type JsonString struct {
	context context.Context
	data    any
}

func (r JsonString) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	{{if and .Jsonv2 (not .HasResponseKey)}}	
		return json.MarshalWrite(w,r.data)
	{{else}}
		{{if .Jsonv2}}
		var err error
		buf := new(bytes.Buffer)
		// MarshalWrite 比 Marshal 更省内存，因为它减少了大切片的扩容分配
		if err = json.MarshalWrite(buf, r.data); err != nil {
			return err
		}
		jsonBytes := buf.Bytes()
		{{else}}
		jsonBytes, err := json.Marshal(r.data)
		if err != nil {
				return err
			}
		{{end}}

		{{if .HasResponseKey}}
		v := r.context.Value(responseKey)
		if v != nil {
			*(v.(*string)) = string(jsonBytes)
		}
		{{end}}
		_, err = w.Write(jsonBytes)
		return err
	{{end}}
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
`

func generateCommon() {
	if commongened {
		return
	}
	var file *astinfo.GenedFile
	file = astinfo.CreateGenedFile("build_in_common")
	file.GetImport(astbasic.SimplePackage("github.com/gin-gonic/gin", "gin"))
	commongened = true
	var content strings.Builder
	Project := astinfo.GlobalProject

	// 准备模板数据
	data := struct {
		HasResponseKey bool
		ImportName     string
		ResponseKey    string
		Jsonv2         bool
	}{}

	if Project.Cfg.Generation.ResponseKey != "" {
		data.HasResponseKey = true
		oneImport := file.GetImport(astbasic.SimplePackage(Project.Cfg.Generation.ResponseMod, "xx"))
		data.ImportName = oneImport.Name
		data.ResponseKey = Project.Cfg.Generation.ResponseKey
	}
	file.GetImport(astbasic.SimplePackage("context", "context"))
	data.Jsonv2 = Project.Cfg.Generation.Jsonv2
	if data.Jsonv2 {
		file.GetImport(astbasic.SimplePackage("encoding/json/v2", "json"))
		if data.HasResponseKey {
			// 如果需要使用 responseKey, it will use bytes.Buffer to marshal the data
			// so we need to import bytes
			// otherwise, it will MarshalWrite directly to writer so no bytes needed
			file.GetImport(astbasic.SimplePackage("bytes", "bytes"))
		}
	} else {
		file.GetImport(astbasic.SimplePackage("encoding/json", "json"))
	}
	file.GetImport(astbasic.SimplePackage("net/http", "http"))

	// 解析并执行模板
	tpl, err := template.New("common").Parse(cJsonTemplate)
	if err != nil {
		// 处理模板解析错误
		panic(err)
	}
	if err := tpl.Execute(&content, data); err != nil {
		// 处理模板执行错误
		panic(err)
	}

	file.AddBuilder(&content)
	file.Save()
}
