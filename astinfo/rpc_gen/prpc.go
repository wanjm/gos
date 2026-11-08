package rpcgen

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/astinfo"
	"github.com/wanjm/gos/astinfo/callable_gen"
)

type PrpcGen struct {
}

func (prpc *PrpcGen) GetName() string {
	return "prpc"
}

func (prpc *PrpcGen) Generate(class *astinfo.Interface, file *astinfo.GenedFile) error {
	var sb strings.Builder
	className := class.InterfaceName + "Struct"
	sb.WriteString("type " + className + " struct {\nclient PrpcClient\n}\n")
	file.AddBuilder(&sb)
	// 生成rpc strutct 代码；
	for _, servlet := range class.Methods {
		prpc.genRpcClientCode(file, className, servlet)
	}
	return nil
}

// ConstructExpr builds the prpc client value for one interface.
// It returns &NameStruct{...} for assignment to the interface variable.
func (prpc *PrpcGen) ConstructExpr(iface *astinfo.Interface, file *astinfo.GenedFile) string {
	host := clientHostExpr(iface, file)
	impt := file.GetImport(&iface.GoSource.Pkg.PkgBasic)
	headerFunc, err := prpcHeaderExpr(iface.Comment.Header, impt.Name)
	if err != nil {
		fmt.Printf("%s on %s\n", err.Error(), iface.InterfaceName)
		os.Exit(1)
	}
	header := ""
	if headerFunc != "" {
		header = "\n\t\t\theader: " + headerFunc + ","
	}
	return fmt.Sprintf(`&%sStruct{
		client: PrpcClient{
			RestfulClient: RestfulClient{
				Prefix: %s,
				rpcLogger: &rpclogger,%s
			},
		},
	}`, iface.InterfaceName, host, header)
}

// prpcHeaderExpr turns a prpc header= name into a function value in the interface package.
// It returns an empty string when header is unset, and an error when the value is a call.
func prpcHeaderExpr(header, importName string) (string, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", nil
	}
	if strings.ContainsAny(header, "()") {
		return "", fmt.Errorf("prpc header %s must be a function name, not a call", header)
	}
	return importName + "." + header, nil
}

// 修改genRpcClientCode函数为使用template的形式
func (prpc *PrpcGen) genRpcClientCode(file *astinfo.GenedFile, structName string, method *astinfo.InterfaceField) {
	// 定义模板字符串
	const clientTemplate = `
	func (receiver *{{.StructName}}) {{.MethodName}}(ctx context.Context, {{.Params}}) ({{.Results}}) {
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
}
`

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
	file.GetImport(astbasic.SimplePackage("context", "context"))
	file.GetImport(astbasic.SimplePackage("errors", "errors"))
	if data.HasResults {
		file.GetImport(astbasic.SimplePackage("encoding/json", "json"))
	}

	// 将生成的代码添加到文件
	file.AddBuilder(&sb)
}

var generated bool

func (prpc *PrpcGen) GenerateCommon(file *astinfo.GenedFile) {
	if generated {
		return
	}
	GenRpcCommon()
	callable_gen.GenBasicError(file)
	generated = true
	file.GetImport(astbasic.SimplePackage("bytes", "bytes"))
	file.GetImport(astbasic.SimplePackage("encoding/json", "json"))
	file.GetImport(astbasic.SimplePackage("context", "context"))
	var content strings.Builder
	content.WriteString(`
type PrpcClient struct {
	RestfulClient
}

// SendRequest calls a prpc method through RestfulClient.
// It posts the argument array and decodes {c,o} from the body for any HTTP status.
func (client *PrpcClient) SendRequest(ctx context.Context, name string, array []any) RpcResult {
	url := client.Prefix + name
	content, marError := json.Marshal(array)
	if marError != nil {
		client.rpcLogger.LogError(ctx, url, marError.Error())
		return RpcResult{C: 1, O: [2]any{nil, &json.RawMessage{}}}
	}
	_, responseBody, err := client.Exchange(ctx, "POST", name, content, true)
	if err != nil {
		return RpcResult{C: 1, O: [2]any{&Error{Message: "send request failed"}, &json.RawMessage{}}}
	}
	var res = RpcResult{
		O: [2]any{&Error{}, &json.RawMessage{}},
	}
	dec := json.NewDecoder(bytes.NewReader(responseBody))
	_ = dec.Decode(&res)
	return res
}
`)
	file.AddBuilder(&content)
}

// InitClientVariable no longer assigns a package global for prpc.
// The client value is produced by ConstructExpr inside initVariable.
func (prpc *PrpcGen) InitClientVariable(map[*astinfo.Interface]*astinfo.VarField, *astinfo.GenedFile) string {
	return ""
}
