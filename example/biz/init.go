package biz

import (
	"context"
	"net/http"

	"github.com/wan_jm/servlet_example/basic"
	"gorm.io/gorm"
)

// @goservlet type=initiator
func GetSql() *HelloRequest {
	return nil
}

// @goservlet type=initiator
func GetSql2(a HelloRequest) *gorm.DB {
	return nil
}

// @goservlet type=filter; group=servlet
func Filter(ctx context.Context, req **http.Request) (res basic.Error) {
	return
}

// @goservlet type=filter; group=servlet
func Filter2(ctx context.Context, req **http.Request) (res basic.Error) {
	return
}
