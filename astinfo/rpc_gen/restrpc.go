package rpcgen

import (
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/astinfo"
)

type RestrpcGen struct {
}

func (r *RestrpcGen) GetName() string {
	return "restrpc"
}

func (r *RestrpcGen) Generate(class *astinfo.Interface, file *astinfo.GenedFile) error {
	if len(class.Methods) == 0 {
		return nil
	}
	var sb strings.Builder
	className := class.InterfaceName + "Struct"
	sb.WriteString("type " + className + " struct {\nclient RestrpcClient\n}\n")
	file.AddBuilder(&sb)
	for _, method := range class.Methods {
		r.genRpcClientCode(file, className, method)
	}
	return nil
}

func (r *RestrpcGen) genRpcClientCode(file *astinfo.GenedFile, structName string, method *astinfo.InterfaceField) {
	const clientTemplate = `
func (receiver *{{.StructName}}) {{.MethodName}}(ctx context.Context, {{.Params}}) ({{.Results}}) {
	err = receiver.client.SendRequest(ctx, "{{.HttpMethod}}", {{.Url}}, {{.RequestArg}}, {{.ResultArg}})
	return
}
`

	httpMethod := method.Comment.Method
	if httpMethod == "" {
		httpMethod = "POST"
	}

	data := struct {
		StructName string
		MethodName string
		Params     string
		Results    string
		RequestArg string
		ResultArg  string
		Url        string
		HttpMethod string
	}{
		StructName: structName,
		MethodName: method.Name,
		Url:        method.Comment.Url,
		HttpMethod: httpMethod,
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
		data.ResultArg = "&obj"
	} else {
		data.ResultArg = "nil"
	}
	results = append(results, "err error")
	data.Results = strings.Join(results, ",")

	tpl, err := template.New("restrpcClient").Parse(clientTemplate)
	if err != nil {
		panic("Failed to parse restrpc client template: " + err.Error())
	}

	var sb strings.Builder
	if err := tpl.Execute(&sb, data); err != nil {
		panic("Failed to execute restrpc client template: " + err.Error())
	}

	file.GetImport(astbasic.SimplePackage("context", "context"))
	file.AddBuilder(&sb)
}

var restrpcGenerated bool

func (r *RestrpcGen) GenerateCommon(file *astinfo.GenedFile) {
	if restrpcGenerated {
		return
	}
	GenRpcCommon()
	restrpcGenerated = true
	file.GetImport(astbasic.SimplePackage("encoding/json", "json"))
	file.GetImport(astbasic.SimplePackage("fmt", "fmt"))
	file.GetImport(astbasic.SimplePackage("net/http", "http"))
	file.GetImport(astbasic.SimplePackage("net/url", "url"))
	file.GetImport(astbasic.SimplePackage("context", "context"))
	file.GetImport(astbasic.SimplePackage("strings", "strings"))
	var content strings.Builder
	content.WriteString(`
// RestError is the object returned by a restrpc call when HTTP status indicates failure.
type RestError struct {
	StatusCode int             "json:\"-\""
	Code       int             "json:\"code\""
	Message    string          "json:\"message\""
	ErrorText  string          "json:\"error\""
	Raw        json.RawMessage "json:\"-\""
}

func (e *RestError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.ErrorText != "" {
		return e.ErrorText
	}
	if len(e.Raw) > 0 {
		return string(e.Raw)
	}
	return fmt.Sprintf("http status %d", e.StatusCode)
}

func (e *RestError) GetErrorCode() int {
	if e.Code != 0 {
		return e.Code
	}
	return e.StatusCode
}

type RestrpcClient struct {
	RestfulClient
}

// SendRequest calls a RESTful method through RestfulClient.
// It turns a non-2xx status into *RestError and decodes a 2xx body into result.
func (client *RestrpcClient) SendRequest(ctx context.Context, method, name string, parameter any, result any) error {
	path := name
	var request any
	if method == http.MethodGet || method == http.MethodHead {
		if parameter != nil {
			query, qerr := restrpcEncodeQuery(parameter)
			if qerr != nil {
				client.rpcLogger.LogError(ctx, client.Prefix+name, qerr.Error())
				return qerr
			}
			if query != "" {
				if strings.Contains(client.Prefix+name, "?") {
					path = name + "&" + query
				} else {
					path = name + "?" + query
				}
			}
		}
	} else {
		request = parameter
	}
	status, responseBody, err := client.Exchange(ctx, method, path, request)
	if err != nil {
		return err
	}
	reqURL := client.Prefix + path
	if status < 200 || status >= 300 {
		restErr := &RestError{
			StatusCode: status,
			Raw:        append(json.RawMessage(nil), responseBody...),
		}
		_ = json.Unmarshal(responseBody, restErr)
		client.rpcLogger.LogError(ctx, reqURL, restErr.Error())
		return restErr
	}
	if result != nil && len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, result); err != nil {
			client.rpcLogger.LogError(ctx, reqURL, err.Error())
			return err
		}
	}
	return nil
}

func restrpcEncodeQuery(parameter any) (string, error) {
	content, err := json.Marshal(parameter)
	if err != nil {
		return "", err
	}
	var m map[string]any
	if err := json.Unmarshal(content, &m); err != nil {
		return "", err
	}
	values := url.Values{}
	for k, v := range m {
		if v == nil {
			continue
		}
		values.Set(k, fmt.Sprint(v))
	}
	return values.Encode(), nil
}
`)
	file.AddBuilder(&content)
}

// ConstructExpr builds the restrpc client value for one interface.
// It returns &NameStruct{...} for assignment to the interface variable.
func (r *RestrpcGen) ConstructExpr(iface *astinfo.Interface, file *astinfo.GenedFile) string {
	host := clientHostExpr(iface, file)
	return fmt.Sprintf(`&%sStruct{
		client: RestrpcClient{
			RestfulClient: RestfulClient{
				Prefix: %s,
				rpcLogger: &rpclogger,
			},
		},
	}`, iface.InterfaceName, host)
}

func (r *RestrpcGen) InitClientVariable(rpcClientVar map[*astinfo.Interface]*astinfo.VarField, file *astinfo.GenedFile) string {
	rpcClientTpl := `
func initRestrpcClient() {
	{{if .HasLogger}}
	var rpclogger {{.LoggerImport}}.{{.LoggerKey}}
	{{else}}
	var rpclogger defaultRpcLogger
	{{end}}

	{{range .RpcFields}}
	{{.ImportName}}.{{.FieldName}} = &{{.TypeName}}Struct{
		client: RestrpcClient{
			RestfulClient: RestfulClient{
				Prefix: {{.Host}},
				rpcLogger: &rpclogger,
			},
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

	tpl, err := template.New("restrpcClientInit").Parse(rpcClientTpl)
	if err != nil {
		log.Fatalf("Failed to parse restrpc client template: %v", err)
	}

	var content strings.Builder
	if err := tpl.Execute(&content, data); err != nil {
		log.Fatalf("Failed to execute restrpc client template: %v", err)
	}

	file.AddBuilder(&content)
	return "initRestrpcClient"
}
