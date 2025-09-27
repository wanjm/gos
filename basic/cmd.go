package basic

type argument struct {
	SourcePath  string // 源码工作忙碌
	GoMod       string // 本项目的module名字
	ModName     string // 数据库的模块名 字符串数组
	SqlDBName   string // 数据库的名字； all表示所有
	MongoDBName string // mongo数据库的名字； all表示所有
	MongoPath   string
}

var Argument argument
