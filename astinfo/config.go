package astinfo

type Generation struct {
	ResponseKey string
	ResponseMod string
	AutoGen     bool
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
