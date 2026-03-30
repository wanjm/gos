package rpcgen

import (
	"strings"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/astinfo"
	"github.com/wanjm/gos/astinfo/callable_gen"
)

var rpcCommonGenerated bool

// GenRpcCommon generates shared rpc infrastructure (rpcLogger, defaultRpcLogger)
// into rpc_common.gen.go. TraceIdNameInContext is generated in build_in_common.
// Called by both PrpcGen and SrpcGen to avoid duplication.
func GenRpcCommon() {
	if rpcCommonGenerated {
		return
	}
	rpcCommonGenerated = true

	// Ensure build_in_common (with TraceIdNameInContext) is generated first
	callable_gen.GenerateBuildInCommon()

	file := astinfo.CreateGenedFile("rpc_common")
	file.GetImport(astbasic.SimplePackage("context", "context"))
	file.GetImport(astbasic.SimplePackage("fmt", "fmt"))

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
`)
	file.AddBuilder(&content)
	file.Save()
}
