package biz

import (
	"context"

	"github.com/wan_jm/servlet_example/basic"
	"gorm.io/gorm"
)

// @goservlet type=servlet;url="/example"
type Hello struct {
	Mysql *gorm.DB
}

// @goservlet url="/hello";
func (hello *Hello) SayHello(ctx context.Context, req *HelloRequest) (res HelloResponse, err basic.Error) {
	res.Greeting = "hello " + req.Name
	return
}

type HelloRequest struct {
	Name string `json:"name"`
}
type HelloResponse struct {
	Greeting string `json:"greeting"`
}
