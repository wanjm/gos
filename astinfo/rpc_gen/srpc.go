package rpcgen

import (
	"fmt"
	"log"
	"sort"
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
	var sb strings.Builder
	className := class.InterfaceName + "Struct"
	sb.WriteString("type " + className + " struct {\nclient SrpcClient\n}\n")
	file.AddBuilder(&sb)
	for _, method := range class.Methods {
		srpc.genRpcClientCode(file, className, method)
	}
	srpc.genWithPrefix(file, class, className)
	return nil
}

// ConstructExpr builds the srpc client value for one interface.
// It returns &NameStruct{...} for assignment to the interface variable.
func (srpc *SrpcGen) ConstructExpr(iface *astinfo.Interface, file *astinfo.GenedFile) string {
	host := clientHostExpr(iface, file)
	return fmt.Sprintf(`&%sStruct{
		client: SrpcClient{
			RestfulClient: RestfulClient{
				Prefix: %s,
				rpcLogger: &rpclogger,
			},
		},
	}`, iface.InterfaceName, host)
}

// genWithPrefix 为 SRPC 客户端生成 WithPrefix。
// 它写入一个复制 logger、改写 Prefix 的方法。
func (srpc *SrpcGen) genWithPrefix(file *astinfo.GenedFile, class *astinfo.Interface, structName string) {
	const withPrefixTemplate = `
// WithPrefix 返回指向 prefix 的同接口客户端。
// 它复制 logger 并改写 Prefix，不修改原客户端。
func (receiver *{{.StructName}}) WithPrefix(prefix string) {{.IfaceType}} {
	cloned := *receiver
	cloned.client.Prefix = prefix
	return &cloned
}
`
	data := struct {
		StructName string
		IfaceType  string
	}{
		StructName: structName,
		IfaceType:  class.RefName(file),
	}
	tpl, err := template.New("srpcWithPrefix").Parse(withPrefixTemplate)
	if err != nil {
		panic("Failed to parse srpc WithPrefix template: " + err.Error())
	}
	var sb strings.Builder
	if err := tpl.Execute(&sb, data); err != nil {
		panic("Failed to execute srpc WithPrefix template: " + err.Error())
	}
	file.AddBuilder(&sb)
}

func (srpc *SrpcGen) genRpcClientCode(file *astinfo.GenedFile, structName string, method *astinfo.InterfaceField) {
	const clientTemplate = `
func (receiver *{{.StructName}}) {{.MethodName}}(ctx context.Context, {{.Params}}) ({{.Results}}) {
	err = receiver.client.SendRequest(ctx, {{.Url}}, {{.RequestArg}}, {{.ResultArg}})
	return
}
`

	data := struct {
		StructName string
		MethodName string
		Params     string
		Results    string
		RequestArg string
		ResultArg  string
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
		data.ResultArg = "&obj"
	} else {
		data.ResultArg = "&struct{}{}"
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
	file.GetImport(astbasic.SimplePackage("context", "context"))
	var content strings.Builder
	content.WriteString(`
// SrpcError is the JSON body of an srpc call.
// A non-zero Code is returned directly as the call error.
type SrpcError struct {
	Code    int    "json:\"code\""
	Message string "json:\"msg\""
	Object  any    "json:\"obj\""
}

// Error returns the srpc failure text.
// It reads msg from the response.
func (e SrpcError) Error() string {
	return e.Message
}

// GetErrorCode returns the srpc business code.
// It reads code from the response.
func (e SrpcError) GetErrorCode() int {
	return e.Code
}

type SrpcClient struct {
	RestfulClient
}

// SendRequest calls a servlet-style method through RestfulClient.
// It posts the request object and decodes {code,msg,obj} from the body for any HTTP status.
func (client *SrpcClient) SendRequest(ctx context.Context, name string, parameter any, result any) error {
	url := client.Prefix + name
	content, marError := json.Marshal(parameter)
	if marError != nil {
		client.rpcLogger.LogError(ctx, url, marError.Error())
		return marError
	}
	_, responseBody, err := client.Exchange(ctx, "POST", name, content, true)
	if err != nil {
		return err
	}
	var res SrpcError
	res.Object = result
	dec := json.NewDecoder(bytes.NewReader(responseBody))
	_ = dec.Decode(&res)
	if res.Code != 0 {
		return res
	}
	return nil
}
`)
	file.AddBuilder(&content)
}

// InitClientVariable assigns srpc WithPrefix globals when they are declared.
// The client value itself is produced by ConstructExpr inside initVariable.
func (srpc *SrpcGen) InitClientVariable(rpcClientVar map[*astinfo.Interface]*astinfo.VarField, file *astinfo.GenedFile) string {
	rpcClientTpl := `
func initSrpcClient() {
	{{range .RpcFields}}
	{{.ImportName}}.{{.FieldName}}WithPrefix = func(prefix string) {{.ImportName}}.{{.TypeName}} {
		return &{{.TypeName}}Struct{
			client: SrpcClient{
				RestfulClient: RestfulClient{
					Prefix: prefix,
					rpcLogger: &rpclogger,
				},
			},
		}
	}
	{{end}}
}
	`
	type RpcFieldData struct {
		ImportName string
		FieldName  string
		TypeName   string
	}

	var fields []RpcFieldData
	for iface, field := range rpcClientVar {
		if iface.GoSource.Pkg.GlobalVar[field.Name+"WithPrefix"] == nil {
			continue
		}
		impt := file.GetImport(&iface.GoSource.Pkg.PkgBasic)
		fields = append(fields, RpcFieldData{
			ImportName: impt.Name,
			FieldName:  field.Name,
			TypeName:   iface.InterfaceName,
		})
	}
	if len(fields) == 0 {
		return ""
	}
	sort.Slice(fields, func(i, j int) bool {
		if fields[i].ImportName != fields[j].ImportName {
			return fields[i].ImportName < fields[j].ImportName
		}
		return fields[i].FieldName < fields[j].FieldName
	})

	tpl, err := template.New("srpcClientInit").Parse(rpcClientTpl)
	if err != nil {
		log.Fatalf("Failed to parse srpc client template: %v", err)
	}

	var content strings.Builder
	if err := tpl.Execute(&content, struct {
		RpcFields []RpcFieldData
	}{RpcFields: fields}); err != nil {
		log.Fatalf("Failed to execute srpc client template: %v", err)
	}

	file.AddBuilder(&content)
	return "initSrpcClient"
}
