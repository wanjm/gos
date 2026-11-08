package rpcgen

import (
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

// RestfulClient is the shared HTTP caller for prpc, srpc, and restrpc.
type RestfulClient struct {
	Prefix    string
	rpcLogger rpcLogger
}

// Exchange sends one HTTP call shared by prpc, srpc, and restrpc.
// It returns the status and body. A non-2xx status is not an error. requireTrace panics when the trace id is missing.
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
