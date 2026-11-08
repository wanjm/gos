package rpcgen

import (
	"fmt"
	"strings"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/astinfo"
	"github.com/wanjm/gos/astinfo/callable_gen"
)

var rpcCommonGenerated bool

// GenRpcCommon generates shared rpc infrastructure into rpc_common.gen.go.
// It writes the logger and RestfulClient.Exchange, which prpc, srpc, and restrpc call.
func GenRpcCommon() {
	if rpcCommonGenerated {
		return
	}
	rpcCommonGenerated = true

	// Ensure build_in_common (with TraceIdNameInContext) is generated first
	callable_gen.GenerateBuildInCommon()

	file := astinfo.CreateGenedFile("rpc_common")
	file.GetImport(astbasic.SimplePackage("bytes", "bytes"))
	file.GetImport(astbasic.SimplePackage("context", "context"))
	file.GetImport(astbasic.SimplePackage("fmt", "fmt"))
	file.GetImport(astbasic.SimplePackage("io", "io"))
	file.GetImport(astbasic.SimplePackage("net/http", "http"))

	var content strings.Builder
	content.WriteString(`
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
` + rpcLoggerVar(file) + `
// RestfulClient is the shared HTTP caller for prpc, srpc, and restrpc.
type RestfulClient struct {
	Prefix    string
	rpcLogger rpcLogger
	// header adds request headers for this client. prpc sets the developer function named by header=.
	header func(context.Context) map[string]string
}

// Exchange sends one HTTP call shared by prpc, srpc, and restrpc.
// It copies client headers onto the request and returns the status and body. A non-2xx status is not an error. requireTrace panics when the trace id is missing.
func (client *RestfulClient) Exchange(ctx context.Context, method, name string, body []byte, requireTrace bool) (int, []byte, error) {
	reqURL := client.Prefix + name
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, reqURL, bodyReader)
	var resp *http.Response
	if err == nil {
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if client.header != nil {
			for key, value := range client.header(ctx) {
				req.Header.Set(key, value)
			}
		}
		if requireTrace {
			req.Header.Set(TraceId, ctx.Value(TraceIdNameInContext).(string))
		} else if v := ctx.Value(TraceIdNameInContext); v != nil {
			if tid, ok := v.(string); ok && tid != "" {
				req.Header.Set(TraceId, tid)
			}
		}
		resp, err = http.DefaultClient.Do(req)
		client.rpcLogger.LogRequest(ctx, reqURL, string(body))
	}
	if err != nil {
		client.rpcLogger.LogError(ctx, reqURL, err.Error())
		return 0, nil, err
	}
	responseBody, _ := io.ReadAll(resp.Body)
	client.rpcLogger.LogResponse(ctx, reqURL, string(responseBody))
	resp.Body.Close()
	return resp.StatusCode, responseBody, nil
}
`)
	file.AddBuilder(&content)
	file.Save()
}

// rpcLoggerVar declares the package-level logger wired clients take the address of.
// It uses the configured logger type, or defaultRpcLogger when none is configured.
func rpcLoggerVar(file *astinfo.GenedFile) string {
	cfg := astinfo.GlobalProject.Cfg.Generation
	if cfg.RpcLoggerKey == "" {
		return "\nvar rpclogger defaultRpcLogger\n"
	}
	impt := file.GetImport(astbasic.SimplePackage(cfg.RpcLoggerMod, "xx"))
	return fmt.Sprintf("\nvar rpclogger %s.%s\n", impt.Name, cfg.RpcLoggerKey)
}

// clientHostExpr returns the Prefix expression for an rpc client.
// A quoted host is copied as a string; a bare name is called in the interface package.
func clientHostExpr(iface *astinfo.Interface, file *astinfo.GenedFile) string {
	impt := file.GetImport(&iface.GoSource.Pkg.PkgBasic)
	host := iface.Comment.Host
	if !strings.HasPrefix(host, `"`) {
		host = impt.Name + "." + host
	}
	return host
}
