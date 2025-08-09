package biz

import (
	"context"
)

// @gos type=servlet;url="/example"
type Hello struct{}

// @gos url="/hello"; title="hello"
func (hello *Hello) SayHello(ctx context.Context, req *HelloRequest) (string, error) {
	return "hello " + req.Name, nil
}

type HelloRequest struct {
	Name string `json:"name"`
}
