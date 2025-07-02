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
	Security    = "security"
	ConstMethod = "method"
	Title       = "title" //定义函数的描述描述，供swagger使用
	Type        = "type"
	Group       = "group"
	//desperate
	Servlet = "servlet" //用于定义struct是servlet，所以默认groupName是servlets
	Prpc    = "prpc"    //用于定义struct是prpc，所以默认groupName是prpc
)
