package basic

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
	InitMain    string // 改为字符串类型，存储模块名称
	Generation  Generation
	SwaggerCfg  SwaggerCfg
	MysqlGenCfg []*DBConfig
}

type DBConfig struct {
	DSN          string
	DBName       string
	DBType       string // 数据库类型，mysql或mongo
	MysqlGenCfgs []*MysqlGenCfg
}
type MysqlGenCfg struct {
	TableNames []string
	RecordIds  []string // 记录id的字段名，生成mongo的结构体；不要时可以为空，或者将所有的不要的dbname放在最后，可以不写
	OutPath    string
	ModulePath string
}

type SwaggerCfg struct {
	ProjectId     int    // 项目id
	ServletFolder int    // 生成的servlet文件夹
	SchemaFolder  int    // 生成的schema文件夹
	UrlPrefix     string // url前缀, 正式环境和本地的路径不一样
	Token         string
}

var Cfg Config

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
