package biz

import (
	"context"
)

// BizHello is injected into Hello. autogen is the deprecated spelling of auto.
// @gos autogen
type BizHello struct {
	Names  map[string]string
	Chanel chan string
}

// @gos type=servlet;url="/example"
type Hello struct {
	BizHello *BizHello
}

// @gos url="/hello"; title="hello"
func (hello *Hello) SayHello(ctx context.Context, req *HelloRequest) (string, error) {
	return "hello " + req.Name, nil
}

type HelloRequest struct {
	Name string `json:"name"`
}
