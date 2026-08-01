# key
1. url  //定义服务路径
2. title //定义服务标题；
3. auto // DI 自动：修饰 struct 则生成注入对象，修饰全局 var 则从 DI 赋值；见下方 `auto`
4. group // 服务所属的群组，修饰struct，和filter；
5. tblName // 修饰entity的tableName
6. dbVariable // 修饰entity的dbVariable， dal中链接驱动的接入；
7. entity // schema 结构上声明：目标实体类型名（同包或可解析的包），用于生成 `FromEntity` 等转换代码；取值 Go 类型名字符串
8. arrays // 逗号分隔的 DB 列名,不是结构体的变量名（与字段 `DbColumnName` / gorm column 对齐）。在生成的 `XxxList` 上为每个列生成 `Get<Field>List() []T` 辅助方法
9. maps // 逗号分隔的 DB 列名。在 `XxxList` 上为每个列生成 `GetMapBy<Field>() map[T]*Entity`（按该字段取值建索引）
10. host //prpc，srpc客户端，提供获取host的地址函数；
11. type // 类型；
12. method //方法；
13. filters //过滤器函数名；
14. header； // function支持header="X-AppId:appId,X-Timestamp:时间戳秒,X-Nonce:随机字符串,X-Sign:签名"； 结构体变量支持header；
15. swagger // swagger=false 时不生成 OpenAPI/Swagger 文档；可在 struct 或 method 上使用
16. enum //添加在const的后面， 表示这是一组取值， 会被编译为enum到前端；
17. dispalyWord 对于const的每个常量，前端的显示值；如果不存在则为变量后面的注释；
18. displayKey 对于const常量，前端多语言时使用的key，默认规则就是名字；

## auto
裸 key（无取值）。含义由修饰目标决定：
1. 修饰 **struct**：自动生成对象供注入（原 `autogen`）
2. 修饰全局 **var**：在 `initVariable` 末尾将该类型对应的 initiator/`auto` struct 实例赋值给该变量（供 filter 等无法注入的场景使用）

### deprecated（保留到下一个大版本，之后删除）
1. `autogen` — 等价于 struct 上的 `auto`

## type
表示struct，method，function的类型；
1. servlet; （类是servlet）；
2. prpc；
    - struct prpc服务端
    - interface 表示prpc客户端
3. srpc； （client是srpc客户端，调用对端 servlet 风格接口：POST JSON，响应 `{code,msg,obj}`）
4. filter； 表示是filter函数；
5. initiator； 初始化函数，返回值供依赖注入；

## srpc
定义 **interface + 全局 var**：gos 生成 `XxxStruct` 实现与 `initSrpcClient`，在 init 时把客户端赋给该 var。

1. interface：`type=srpc`，`host` 为 base URL 字符串字面量，或同包内「返回 host 的函数调用」
2. 每个方法：`url` 为相对路径（会拼到 `host` 后面）；签名为 `(ctx, req) (resp, error)`
3. 必须声明 `var Xxx InterfaceType`，否则不会生成注入

```go
// host 函数：返回对端服务前缀（也可直接写 host="http://127.0.0.1:8080"）
func GetBookHost() string {
	return "http://book.internal"
}

// @gos type=srpc; host=GetBookHost()
type BookClient interface {
	// @gos url="/book/get"
	GetBook(ctx context.Context, req *GetBookRequest) (*GetBookResponse, error)

	// @gos url="/book/list"
	ListBook(ctx context.Context, req *ListBookRequest) (*ListBookResponse, error)
}

// 全局 var：gos 生成的 initSrpcClient 会赋值
var BookSvc BookClient
```

业务侧直接调用即可（无需手写 HTTP）：

```go
resp, err := BookSvc.GetBook(ctx, &GetBookRequest{Id: 1})
```

## method
供servlet使用；
1. GET
2. POST

## filters
逗号分隔的函数名。直接是filter函数的名字。 系统会自动从filter中去找；

## FromEntity autogen
为 schema 结构体添加 `@gos entity=EntityName` 注解时，系统会自动为其在当前包下的 `schema.gen.go` 文件中生成：
1. `FromEntity(e *EntityName)` 方法。
2. `FromEntitys(entitys []*EntityName)` （针对当前 schema 结构体的数组类型）。

字段匹配规则：
- 生成器会自动遍历 entity 结构体（支持嵌套结构体，最多向内搜索5层），查找与 schema 中字段名相同的字段。
- 如果找到名称匹配的嵌套字段（例如 schema 中的 `ID` 匹配到了 entity 中 `BaseInfo.ID`），也会自动生成对应的赋值代码 `s.ID = e.BaseInfo.ID`。
- 如果 schema 中某些字段在 entity 内找不到同名字段，编译器会提示 WARNING，此时可以在 schema 结构体上实现 `FormatEntity()` 方法，生成器检测到该方法后，会在生成的 `FromEntity` 内部进行调用，以便补充其他自定义逻辑。