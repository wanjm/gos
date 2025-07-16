package gen

import (
	cors "github.com/gin-contrib/cors"
	gin "github.com/gin-gonic/gin"
	biz "github.com/wan_jm/servlet_example/biz"
	gorm "gorm"
	sync "sync"
)

type Response struct {
	Code    int         "json:\"code\""
	Message string      "json:\"message,omitempty\""
	Object  interface{} "json:\"obj,omitempty\""
}

type Config struct {
	CertFile   string
	KeyFile    string
	Cors       bool
	Addr       string
	ServerName string
}

func getAddr[T any](a T) *T {
	return &a
}

type server struct {
	filters       gin.HandlersChain
	routerInitors []func(*gin.Engine)
}

var servers map[string]*server

func Run(config ...Config) *sync.WaitGroup {
	prepare()
	var wg sync.WaitGroup
	for _, c := range config {
		wg.Add(1)
		go run(&wg, c)
	}
	return &wg
}
func run(wg *sync.WaitGroup, config Config) {
	var router *gin.Engine = gin.New()
	router.ContextWithFallback = true
	if config.Cors {
		config := cors.DefaultConfig()
		config.AllowAllOrigins = true
		config.AllowHeaders = append(config.AllowHeaders, "*")
		router.Use(cors.New(config))
	}
	//如果不存在，则启动就失败，不需要检查
	server := servers[config.ServerName]
	if server.filters != nil {
		router.Use(server.filters...)
	}
	for _, routerInitor := range server.routerInitors {
		routerInitor(router)
	}
	if config.CertFile != "" {
		router.RunTLS(config.Addr, config.CertFile, config.KeyFile)
	} else {
		router.Run(config.Addr)
	}
	wg.Done()
}

const TraceId = "TraceId"

var (
	__global__0 *biz.HelloRequest
	__global__1 *gorm.DB
)

func initVariable() {
	__global__0 = biz.GetSql()
	__global__1 = biz.GetSql2(*__global__0)
}
func Prepare() {
	initVariable()

}
func prepare() {
	Prepare()
}
