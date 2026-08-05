package astinfo

const (
	NOUSAGE = iota
	CREATOR
	SERVLET
	PRPC
	INITIATOR
	FILTER
	WEBSOCKET
	TITLE
)

const (
	UrlFilter   = "urlfilter"
	Url         = "url" // 定义函数为servlet，默认method为POST
	Creator     = "creator"
	Initiator   = "initiator"
	Websocket   = "websocket"
	FilterConst = "filter"
	UserFilter  = "filters"
	Security    = "security"
	ConstMethod = "method"
	Title          = "title"   //定义函数的描述描述，供swagger使用
	SwaggerKey     = "swagger" // swagger=false 时不生成 OpenAPI 文档
	Type           = "type"
	Group       = "group"
	AutoKey = "auto"    // bare key: struct => gen DI instance; global var => fill from DI
	AutoGen = "autogen" // deprecated alias of auto on struct; remove after next major release
	Host    = "host"    //rpcclient 使用
	tblName     = "tblName"
	dbVarible   = "dbVariable"
	entityKey   = "entity"
	arraysKey   = "arrays"
	mapsKey     = "maps"
	Macro       = "macro"
	//desperate
	Servlet = "servlet" //用于定义struct是servlet，所以默认groupName是servlets
	Prpc    = "prpc"    //用于定义struct是prpc，所以默认groupName是prpc
	Srpc    = "srpc"    //用于定义struct是srpc，server rpc client，调用servlet风格接口
	Restrpc = "restrpc" // REST 客户端：JSON 请求，HTTP status 表示成败，失败时返回服务端对象
)

const (
	CreateTime = "create_time"
	Id         = "id"
	GORM       = "gorm"
	BSON       = "bson"
	JSON       = "json"
	VALIDATE   = "validate"
	HEADER     = "header"
	DEFAULT    = "default"
	VALID      = "valid"
	Column     = "column"
)
