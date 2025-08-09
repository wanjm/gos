package biz

import (
	"context"
	"net/http"

	"github.com/wan_jm/servlet_example/basic"
	"gorm.io/gorm"
)

// @gos type=initiator
func GetSql() *HelloRequest {
	return nil
}

// @gos type=initiator
func GetSql2(a HelloRequest) *gorm.DB {
	return nil
}

// @gos type=filter; group=servlet
func Filter(ctx context.Context, req **http.Request) (res basic.Error) {
	return
}

// @gos type=filter; group=servlet
func Filter2(ctx context.Context, req **http.Request) (res basic.Error) {
	return
}
