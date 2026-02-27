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
	CommonMod    string // github.com/wanjm/common 的别名
	AutoGen      bool
	FlutterPath  string // 自动生成前端的flutter路径
	Jsonv2       bool   // 是否使用jsonv2
}
type Config struct {
	InitMain   string // 改为字符串类型，存储模块名称
	Generation Generation
	SwaggerCfg SwaggerCfg
	DBConfig   []*DBConfig //配置链接的数据库信息
}

type DBConfig struct {
	DSN       string
	DBName    string         //用于生成dal中的db变量名；
	DBType    string         // 数据库类型，mysql或mongo
	DbGenCfgs []*TableGenCfg // 数据库中表的记录；
}

// TableCfg 单个表的生成配置
// Name:     表名或集合名
// Arrays:   需要为实体集合生成返回数组的方法对应字段列表，例如 ["UserId"]
// Maps:     需要为实体集合生成返回 Map 的方法对应字段列表，例如 ["UserId"]
// RecordIds: 在 mongo 中用于抽样文档生成结构体的记录 id 列表
type TableCfg struct {
	Name      string   // table name (DB table/collection name)
	Arrays    []string // e.g. ["UserId"] - generate arrays annotation
	Maps      []string // e.g. ["UserId"] - generate maps annotation
	RecordIds []string // for Mongo: sample document ObjectID hex strings
}

type TableGenCfg struct {
	Tables     []TableCfg // 表配置列表，替代原来的 TableNames/RecordIds
	OutPath    string
	ModulePath string
	DBName     string //从DBConfig中复制，无需填写
}

type SwaggerCfg struct {
	ProjectId     int    // 项目id
	ServletFolder int    // 生成的servlet文件夹
	SchemaFolder  int    // 生成的schema文件夹
	UrlPrefix     string // url前缀, 正式环境和本地的路径不一样
	Token         string
	JsonName      string
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
	generation := &config.Generation
	if generation.CommonMod == "" {
		generation.CommonMod = "github.com/wanjm/common"
	}
	if generation.TraceKey != "" && generation.TraceKeyMod == "" {
		generation.TraceKeyMod = "github.com/wanjm/common/trace"
	}
	if generation.ResponseKey != "" && generation.ResponseMod == "" {
		generation.ResponseMod = "github.com/wanjm/common/response"
	}
}
