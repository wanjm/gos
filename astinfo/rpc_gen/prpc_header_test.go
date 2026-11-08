package rpcgen

import "testing"

func TestPrpcHeaderExpr(t *testing.T) {
	got, err := prpcHeaderExpr("InitHeader", "prpc")
	if err != nil {
		t.Fatal(err)
	}
	if got != "prpc.InitHeader" {
		t.Fatalf("got %s", got)
	}

	got, err = prpcHeaderExpr("", "prpc")
	if err != nil || got != "" {
		t.Fatalf("empty header: got %q err %v", got, err)
	}

	if _, err = prpcHeaderExpr("InitHeader()", "prpc"); err == nil {
		t.Fatal("a call expression should be rejected")
	}
}
