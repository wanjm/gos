package biz

import (
	"context"
	"net/http"

	"github.com/wan_jm/servlet_example/basic"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// GetSql supplies the HelloRequest other initiators depend on.
// It runs first because those initiators take HelloRequest.
// @gos type=initiator
func GetSql() *HelloRequest {
	return nil
}

// GetSql2 supplies the *gorm.DB injected into fields named Db.
// The result name db matches dbVariable=db on the MySQL entity.
// @gos type=initiator
func GetSql2(a HelloRequest) (db *gorm.DB) {
	return nil
}

// NewPrimary supplies the *gorm.DB injected into fields named Primary.
// It waits for HelloRequest, so it runs after GetSql.
// @gos type=initiator
func NewPrimary(a HelloRequest) (primary *gorm.DB) {
	return nil
}

// NewReplica supplies the *gorm.DB injected into fields named Replica.
// It is a second *gorm.DB, selected by the result name rather than by type alone.
// @gos type=initiator
func NewReplica(a HelloRequest) (replica *gorm.DB) {
	return nil
}

// NewMongo supplies the *mongo.Database injected into fields named MongoDB.
// The result name mongoDB matches dbVariable=mongoDB on the Mongo entity.
// @gos type=initiator
func NewMongo() (mongoDB *mongo.Database) {
	return nil
}

// Filter is servlet-group middleware with an empty url, so it wraps every servlet route.
// It returns nil so the request continues.
// @gos type=filter; group=servlet
func Filter(ctx context.Context, req **http.Request) (res *basic.Error) {
	return
}

// Filter2 is a second global servlet filter. group must match the servlet server.
// It returns nil so the request continues.
// @gos type=filter; group=servlet
func Filter2(ctx context.Context, req **http.Request) (res *basic.Error) {
	return
}

// SwaggerHelloHeaderFilter binds to servlet routes whose url contains /hello.
// header is copied into that route's OpenAPI parameters.
// @gos type=filter; group=servlet; url="/hello"; header="X-Example-Header"
func SwaggerHelloHeaderFilter(ctx context.Context, req **http.Request) (res *basic.Error) {
	return
}

// AdminGate is bound only by filters=AdminGate, because its url is not part of the route path.
// It returns nil so the request continues.
// @gos type=filter; group=servlet; url="/gate"
func AdminGate(ctx context.Context, req **http.Request) (res *basic.Error) {
	return
}

// RestfulGuard is global middleware for the restful group only.
// It does not run on servlet or prpc routes.
// @gos type=filter; group=restful
func RestfulGuard(ctx context.Context, req **http.Request) (res *basic.Error) {
	return
}

// PrpcGuard is global middleware for the prpc group.
// An empty url attaches it to every prpc route.
// @gos type=filter; group=prpc
func PrpcGuard(ctx context.Context, req **http.Request) (res *basic.Error) {
	return
}

// PrpcPingFilter attaches to prpc routes whose url contains /ping.
// It returns nil so the ping call continues.
// @gos type=filter; group=prpc; url="/ping"
func PrpcPingFilter(ctx context.Context, req **http.Request) (res *basic.Error) {
	return
}
