package astinfo

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Generation struct {
	TraceKey     string // 用于定义traceKy的结构体名字；用于context中记录traceId
	TraceKeyMod  string // 用于定义traceKy的结构体所在的包名；
	ResponseKey  string // 用于定义Response的结构体名字；用于context中记录一个http请求的Response String
	ResponseMod  string // 用于定义Response的结构体所在的包名；
	RpcLoggerKey string // 用于定义RpcLogger的结构体名字; 用于打印rpc请求的日志
	RpcLoggerMod string // 用于定义RpcLogger的结构体所在的包名；
	AutoGen      bool
}
type Config struct {
	InitMain   bool
	Generation Generation
	SwaggerCfg SwaggerCfg
}
type SwaggerCfg struct {
	ProjectId     string // 项目id
	ServletFolder string // 生成的servlet文件夹
	SchemaFolder  string // 生成的schema文件夹
	UrlPrefix     string // url前缀, 正式环境和本地的路径不一样
	Token         string
}

func (config *Config) Load() {
	buf, err := os.ReadFile("project.public.toml")
	if err == nil {
		_, err = toml.Decode(string(buf), config)
		if err != nil {
			panic(err)
		}
	}
	buf, err = os.ReadFile("project.private.toml")
	if err == nil {
		_, err = toml.Decode(string(buf), config)
		if err != nil {
			panic(err)
		}
	}
}
