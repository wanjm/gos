package rpcgen

import (
	"strings"
	"testing"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/astinfo"
)

func TestPrpcConstructExpr(t *testing.T) {
	pkg := &astinfo.Package{}
	pkg.Name = "prpc"
	pkg.ModPath = "example.com/prpc"
	iface := &astinfo.Interface{
		InterfaceName: "Jsinternal",
		GoSource:      &astinfo.Gosourse{Pkg: pkg},
	}
	iface.Comment.Host = "GetHost()"
	iface.Comment.Header = "InitHeader"
	file := astbasic.CreateGenedFile("wire")
	got := (&PrpcGen{}).ConstructExpr(iface, file)
	if !strings.Contains(got, "&JsinternalStruct{") {
		t.Fatalf("missing struct literal: %s", got)
	}
	if !strings.Contains(got, "Prefix: prpc.GetHost()") {
		t.Fatalf("missing host: %s", got)
	}
	if !strings.Contains(got, "header: prpc.InitHeader") {
		t.Fatalf("missing header: %s", got)
	}
	if !strings.Contains(got, "rpcLogger: &rpclogger") {
		t.Fatalf("missing logger: %s", got)
	}
}

func TestSrpcConstructExprUsesHostCall(t *testing.T) {
	pkg := &astinfo.Package{}
	pkg.Name = "rpc"
	pkg.ModPath = "example.com/rpc"
	iface := &astinfo.Interface{
		InterfaceName: "LiveclassGo",
		GoSource:      &astinfo.Gosourse{Pkg: pkg},
	}
	iface.Comment.Host = `"http://live"`
	file := astbasic.CreateGenedFile("wire")
	got := (&SrpcGen{}).ConstructExpr(iface, file)
	if !strings.Contains(got, "Prefix: \"http://live\"") {
		t.Fatalf("quoted host should be copied: %s", got)
	}
}
