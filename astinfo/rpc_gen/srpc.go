package rpcgen

import (
	"log"
	"strings"
	"text/template"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/astinfo"
)

type SrpcGen struct {
}

func (srpc *SrpcGen) GetName() string {
	return "srpc"
}

func (srpc *SrpcGen) Generate(class *astinfo.Interface, file *astinfo.GenedFile) error {
	if len(class.Methods) == 0 {
		return nil
	}
	var sb strings.Builder
	className := class.InterfaceName + "Struct"
	sb.WriteString("type " + className + " struct {\nclient SrpcClient\n}\n")
	file.AddBuilder(&sb)
	for _, method := range class.Methods {
		srpc.genRpcClientCode(file, className, method)
	}
	return nil
}

func (srpc *SrpcGen) genRpcClientCode(file *astinfo.GenedFile, structName string, method *astinfo.InterfaceField) {
	const clientTemplate = `
func (receiver *{{.StructName}}) {{.MethodName}}(ctx context.Context, {{.Params}}) ({{.Results}}) {
	err = receiver.client.SendRequest(ctx, {{.Url}}, {{.RequestArg}}, &obj)
	return
}
`

	data := struct {
		StructName string
		MethodName string
		Params     string
		Results    string
		RequestArg string
		Url        string
	}{
		StructName: structName,
		MethodName: method.Name,
		Url:        method.Comment.Url,
	}

	var args []string
	var params []string
	for i, l := 1, len(method.Params); i < l; i++ {
		param := method.Params[i]
		info := param.Name + " " + param.Type.RefName(file)
		params = append(params, info)
		args = append(args, param.Name)
	}
	data.Params = strings.Join(params, ",")

	if len(args) == 1 {
		data.RequestArg = args[0]
	} else if len(args) > 1 {
		data.RequestArg = "[]interface{}{" + strings.Join(args, ",") + "}"
	} else {
		data.RequestArg = "nil"
	}

	var results []string
	if len(method.Results) >= 2 {
		resultP0 := method.Results[0]
		results = append(results, "obj "+resultP0.Type.RefName(file))
	}
	results = append(results, "err error")
	data.Results = strings.Join(results, ",")

	tpl, err := template.New("srpcClient").Parse(clientTemplate)
	if err != nil {
		panic("Failed to parse srpc client template: " + err.Error())
	}

	var sb strings.Builder
	if err := tpl.Execute(&sb, data); err != nil {
		panic("Failed to execute srpc client template: " + err.Error())
	}

	file.GetImport(astbasic.SimplePackage("context", "context"))
	file.AddBuilder(&sb)
}

var srpcGenerated bool

func (srpc *SrpcGen) GenerateCommon(file *astinfo.GenedFile) {
	if srpcGenerated {
		return
	}
	GenRpcCommon()
	srpcGenerated = true
	file.GetImport(astbasic.SimplePackage("bytes", "bytes"))
	file.GetImport(astbasic.SimplePackage("encoding/json", "json"))
	file.GetImport(astbasic.SimplePackage("errors", "errors"))
	file.GetImport(astbasic.SimplePackage("net/http", "http"))
	file.GetImport(astbasic.SimplePackage("io", "io"))
	file.GetImport(astbasic.SimplePackage("context", "context"))
	var content strings.Builder
	content.WriteString(`
type SrpcClient struct {
	Prefix    string
	rpcLogger rpcLogger
}

func (client *SrpcClient) SendRequest(ctx context.Context, name string, array any, result any) error {
	url := client.Prefix + name
	content, marError := json.Marshal(array)
	if marError != nil {
		client.rpcLogger.LogError(ctx, url, marError.Error())
		return marError
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
		return err
	}
	requestBody, _ := io.ReadAll(resp.Body)
	client.rpcLogger.LogResponse(ctx, url, string(requestBody))
	resp.Body.Close()
	var res struct {
		Code    int    ` + "`json:\"code\"`" + `
		Message string ` + "`json:\"message\"`" + `
		Object  any    ` + "`json:\"obj\"`" + `
	}
	res.Object = result
	dec := json.NewDecoder(bytes.NewReader(requestBody))
	_ = dec.Decode(&res)
	if res.Code != 0 {
		return errors.New(res.Message)
	}
	return nil
}
`)
	file.AddBuilder(&content)
}

func (srpc *SrpcGen) InitClientVariable(rpcClientVar map[*astinfo.Interface]*astinfo.VarField, file *astinfo.GenedFile) string {
	rpcClientTpl := `
func initSrpcClient() {
	{{if .HasLogger}}
	var rpclogger {{.LoggerImport}}.{{.LoggerKey}}
	{{else}}
	var rpclogger defaultRpcLogger
	{{end}}

	{{range .RpcFields}}
	{{.ImportName}}.{{.FieldName}} = &{{.TypeName}}Struct{
		client: SrpcClient{
			Prefix: {{.Host}},
			rpcLogger: &rpclogger,
		},
	}
	{{end}}
}
	`
	generationCfg := astinfo.GlobalProject.Cfg.Generation
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

	if data.HasLogger {
		data.LoggerImport = file.GetImport(astbasic.SimplePackage(generationCfg.RpcLoggerMod, "xx")).Name
		data.LoggerKey = generationCfg.RpcLoggerKey
	}

	for iface, field := range rpcClientVar {
		impt := file.GetImport(&iface.GoSource.Pkg.PkgBasic)
		host := iface.Comment.Host

		if !strings.HasPrefix(host, `"`) {
			host = impt.Name + "." + host
		}

		data.RpcFields = append(data.RpcFields, RpcFieldData{
			ImportName: impt.Name,
			FieldName:  field.Name,
			TypeName:   iface.InterfaceName,
			Host:       host,
		})
	}

	tpl, err := template.New("srpcClientInit").Parse(rpcClientTpl)
	if err != nil {
		log.Fatalf("Failed to parse srpc client template: %v", err)
	}

	var content strings.Builder
	if err := tpl.Execute(&content, data); err != nil {
		log.Fatalf("Failed to execute srpc client template: %v", err)
	}

	file.AddBuilder(&content)
	return "initSrpcClient"
}
