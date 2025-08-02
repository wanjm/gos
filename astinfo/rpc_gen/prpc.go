package rpcgen

import (
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/wan_jm/servlet/astinfo"
)

type PrpcGen struct {
}

func (prpc *PrpcGen) GetName() string {
	return "prpc"
}

func (prpc *PrpcGen) Generate(class *astinfo.Interface, file *astinfo.GenedFile) error {
	if len(class.Methods) == 0 {
		return nil
	}
	var sb strings.Builder
	className := class.InterfaceName + "Struct"
	sb.WriteString("type " + className + " struct {\nclient RpcClient\n}\n")
	file.AddBuilder(&sb)
	// 生成rpc strutct 代码；
	for _, servlet := range class.Methods {
		prpc.genRpcClientCode(file, className, servlet)
	}
	return nil
}

// 修改genRpcClientCode函数为使用template的形式
func (prpc *PrpcGen) genRpcClientCode(file *astinfo.GenedFile, structName string, method *astinfo.InterfaceField) {
	// 定义模板字符串
	const clientTemplate = `func (receiver *{{.StructName}}) {{.MethodName}}(ctx context.Context, {{.Params}}) ({{.Results}}) {
    var argument = []interface{}{ {{.Args}} }

    var res = receiver.client.SendRequest(ctx, {{.Url}}, argument)
    if res.C != 0 {
        err = errors.New("failed to call {{.MethodName}}")
        return
    }
    if res.O[0] != nil {
        err = res.O[0].(error)
        return 
    }
    {{if .HasResults}}    
    //无论object是否位指针，都需要取地址
    json.Unmarshal(*res.O[1].(*json.RawMessage), &obj)
    {{end}}    return
}`

	// 准备模板数据
	data := struct {
		StructName string
		MethodName string
		Params     string
		Results    string
		Args       string
		Url        string
		HasResults bool
	}{
		StructName: structName,
		MethodName: method.Name,
		Url:        method.Comment.Url,
		HasResults: len(method.Results) >= 2,
	}

	// 处理参数
	var args []string
	var params []string
	for i, l := 1, len(method.Params); i < l; i++ {
		param := method.Params[i]
		info := param.Name + " " + param.Type.RefName(file)
		params = append(params, info)
		args = append(args, param.Name)
	}
	data.Params = strings.Join(params, ",")
	data.Args = strings.Join(args, ",")

	// 处理返回值
	var results []string
	if len(method.Results) >= 2 {
		resultP0 := method.Results[0]
		info := "obj " + resultP0.Type.RefName(file)
		results = append(results, info)
	}
	results = append(results, "err error")
	data.Results = strings.Join(results, ",")

	// 创建并解析模板
	tpl, err := template.New("client").Parse(clientTemplate)
	if err != nil {
		panic("Failed to parse client template: " + err.Error())
	}

	// 渲染模板到strings.Builder
	var sb strings.Builder
	if err := tpl.Execute(&sb, data); err != nil {
		panic("Failed to execute client template: " + err.Error())
	}

	// 添加必要的导入
	file.GetImport(astinfo.SimplePackage("context", "context"))
	file.GetImport(astinfo.SimplePackage("errors", "errors"))
	if data.HasResults {
		file.GetImport(astinfo.SimplePackage("encoding/json", "json"))
	}

	// 将生成的代码添加到文件
	file.AddBuilder(&sb)
}

var generated bool

func (prpc *PrpcGen) GenerateCommon(file *astinfo.GenedFile) {
	if generated {
		return
	}
	generated = true
	file.GetImport(astinfo.SimplePackage("bytes", "bytes"))
	file.GetImport(astinfo.SimplePackage("encoding/json", "json"))
	file.GetImport(astinfo.SimplePackage("fmt", "fmt"))
	file.GetImport(astinfo.SimplePackage("net/http", "http"))
	file.GetImport(astinfo.SimplePackage("io", "io"))
	file.GetImport(astinfo.SimplePackage("context", "context"))
	var content strings.Builder
	content.WriteString(`
type Error struct {
	Code    int    "json:\"code\""
	Message string "json:\"message\""
}

func (error *Error) Error() string {
	return error.Message
}

type RpcResult struct {
	C int    "json:\"c\""
	O [2]any "json:\"o\""
}
type rpcLogger interface {
	LogRequest(ctx context.Context, url, request string)
	LogResponse(ctx context.Context, url, response string)
	LogError(ctx context.Context, url, err string)
}
type defaultRpcLogger struct{}

func (logger *defaultRpcLogger) LogRequest(_ context.Context, url, request string) {
	fmt.Printf("Request to '%s' content='%s'\n", url, request)
}
func (logger *defaultRpcLogger) LogResponse(_ context.Context, url, response string) {
	fmt.Printf("Response of '%s' content='%s'\n", url, response)
}

func (logger *defaultRpcLogger) LogError(_ context.Context, url, err string) {
	fmt.Printf("Error in '%s' err=%s\n", url, err)
}

type RpcClient struct {
	Prefix    string
	rpcLogger rpcLogger
}

func (client *RpcClient) SendRequest(ctx context.Context, name string, array []any) RpcResult {
	url := client.Prefix + name
	content, marError := json.Marshal(array)
	if marError != nil {
		client.rpcLogger.LogError(ctx, url, marError.Error())
		return RpcResult{C: 1, O: [2]any{nil, &json.RawMessage{}}}
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(content))
	var resp *http.Response
	if err == nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(TraceId, ctx.Value(TraceIdNameInContext).(string))
		resp, err = http.DefaultClient.Do(req)
		client.rpcLogger.LogRequest(ctx, url, string(content))
	}
	if err != nil {
		client.rpcLogger.LogError(ctx, url, err.Error())
		return RpcResult{C: 1, O: [2]any{&Error{Message: "send request failed"}, &json.RawMessage{}}}
	}
	requestBody, _ := io.ReadAll(resp.Body)
	client.rpcLogger.LogResponse(ctx, url, string(requestBody))
	resp.Body.Close()
	var res = RpcResult{
		O: [2]any{&Error{}, &json.RawMessage{}},
	}
	dec := json.NewDecoder(bytes.NewReader(requestBody))
	_ = dec.Decode(&res)
	return res
}
`)
	key := astinfo.GlobalProject.Cfg.Generation.TraceKey
	module := astinfo.GlobalProject.Cfg.Generation.TraceKeyMod
	if key != "" {
		// prpc的发送请求是，会向http头添加traceId，需要使用该变量
		oneImport := file.GetImport(astinfo.SimplePackage(module, "xx"))
		content.WriteString(fmt.Sprintf("var TraceIdNameInContext = %s.%s{}\n", oneImport.Name, key))
	} else {
		content.WriteString("var TraceIdNameInContext = \"badTraceIdName plase config in Generation TraceKeyMod\"\n")
	}
	file.AddBuilder(&content)
}

func (prpc *PrpcGen) InitClientVariable(rpcClientVar map[*astinfo.Interface]*astinfo.VarField, file *astinfo.GenedFile) string {
	rpcClientTpl := `
func initRpcClient() {
	{{if .HasLogger}}
	var rpclogger {{.LoggerImport}}.{{.LoggerKey}}
	{{else}}
	var rpclogger defaultRpcLogger
	{{end}}

	{{range .RpcFields}}
	{{.ImportName}}.{{.FieldName}} = &{{.TypeName}}Struct{
		client: RpcClient{
			Prefix: {{.Host}},
			rpcLogger: &rpclogger,
		},
	}
	{{end}}
}
	`
	generationCfg := astinfo.GlobalProject.Cfg.Generation
	// 准备模板数据
	type RpcFieldData struct {
		ImportName string
		FieldName  string
		TypeName   string
		Host       string
	}

	data := struct {
		HasLogger    bool
		LoggerImport string
		LoggerKey    string
		RpcFields    []RpcFieldData
	}{
		HasLogger: generationCfg.RpcLoggerKey != "",
	}

	// 处理日志配置
	if data.HasLogger {
		data.LoggerImport = file.GetImport(astinfo.SimplePackage(generationCfg.RpcLoggerMod, "xx")).Name
		data.LoggerKey = generationCfg.RpcLoggerKey
	}

	for iface, field := range rpcClientVar {
		_ = field
		impt := file.GetImport(iface.Pkg)
		host := iface.Comment.Host

		if !strings.HasPrefix(host, `"`) {
			host = impt.Name + "." + host
		}

		data.RpcFields = append(data.RpcFields, RpcFieldData{
			ImportName: impt.Name,
			FieldName:  "JsinternalClient",
			TypeName:   iface.InterfaceName,
			Host:       host,
		})
	}

	// 解析并执行模板
	tpl, err := template.New("rpcClient").Parse(rpcClientTpl)
	if err != nil {
		log.Fatalf("Failed to parse rpc client template: %v", err)
	}

	var content strings.Builder
	if err := tpl.Execute(&content, data); err != nil {
		log.Fatalf("Failed to execute rpc client template: %v", err)
	}

	file.AddBuilder(&content)
	return ""
}
