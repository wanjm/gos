# gos 使用指南

## 1. 简介

`gos` 用一系列注解生成重复胶水代码，打通前后端联动，让开发者把精力放在业务上。

典型 HTTP 链路里，后端要做注册、路由、反序列化、校验、拦截、调用、序列化；前端要做请求序列化、发请求、解析响应；两边还要对齐 URL、参数格式、返回值与语义。这些对接成本由 `gos` 承担。

---

## 2. 能力概览（我们有什么）

| 类别 | 能力 |
|------|------|
| HTTP / 路由 | URL 绑定、Swagger、filter 绑定（URL / 定义） |
| 数据层 | entity 自动生成（含注释）、列名生成、常用 DAL、MySQL/Mongo 查询差异屏蔽（`Eq` / `Ne` / `In` / `NotIn` / `Gt` / `Gte` / `Lt` / `Lte` / `Exist`） |
| 映射与集合 | entity → schema（`FromEntity`）、list / map 辅助方法 |
| 依赖注入 | 变量生成、自动注入、依赖识别与排序、同类型多对象注入 |
| RPC | 内部 RPC（prpc / srpc / restrpc）定义与调用 |
| 前端联动 | 结构体、API 接口与调用函数、常量、table 生成（含注释） |

**gos 覆盖的三块工作：**

1. 后端服务注解与生成（路由、filter、注入、RPC）
2. 数据库代码自动生成（entity + DAL）
3. 前端代码生成（结构体、API、常量、table）

---

## 3. 详细用法

### 3.1 注解 key 一览

| key | 含义 |
|-----|------|
| `url` | 服务路径 |
| `title` | 服务标题 |
| `auto` | DI：修饰 struct 则生成注入对象；修饰全局 var 则从 DI 赋值（见 [auto](#32-auto)） |
| `group` | 服务所属群组；修饰 struct 与 filter |
| `tblName` | entity 的表名 |
| `dbVariable` | entity 的 DB 变量；DAL 连接驱动接入点 |
| `entity` | schema 上声明目标实体类型名（同包或可解析），用于生成 `FromEntity` 等；取值为 Go 类型名字符串 |
| `arrays` | 逗号分隔的 **DB 列名**（非结构体字段名；与 `DbColumnName` / gorm column 对齐）。在 `XxxList` 上为每列生成 `Get<Field>List() []T` |
| `maps` | 逗号分隔的 DB 列名。在 `XxxList` 上为每列生成 `GetMapBy<Field>() map[T]*Entity` |
| `host` | prpc / srpc / restrpc 客户端：获取 host 的地址函数 |
| `type` | 类型（见 [type](#33-type)） |
| `method` | HTTP method（servlet / restrpc） |
| `filters` | 过滤器函数名（逗号分隔） |
| `header` | function 支持 `header="X-AppId:appId,..."`；结构体变量也可声明 header |
| `swagger` | `swagger=false` 时不生成 OpenAPI；可用于 struct 或 method |
| `enum` | 加在 const 后，表示一组取值，会编译为前端 enum |
| `dispalyWord` | const 每个常量的前端显示值；缺省用变量后注释 |
| `displayKey` | const 前端多语言 key；默认用名字 |

### 3.2 auto

裸 key（无取值）。含义由修饰目标决定：

1. 修饰 **struct**：自动生成对象供注入（原 `autogen`）
2. 修饰全局 **var**：在 `initVariable` 末尾将该类型对应的 initiator / `auto` struct 实例赋给该变量（供 filter 等无法注入的场景）

#### deprecated（保留到下一个大版本，之后删除）

1. `autogen` — 等价于 struct 上的 `auto`

### 3.3 type

表示 struct / method / function 的类型：

| type | 说明 |
|------|------|
| `servlet` | HTTP servlet |
| `prpc` | struct = prpc 服务端；interface = prpc 客户端 |
| `srpc` | 客户端调用对端 servlet 风格接口：POST JSON，响应 `{code,msg,obj}` |
| `restrpc` | 客户端调用对端 RESTful 接口：JSON 请求，HTTP status 表示成败 |
| `filter` | filter 函数 |
| `initiator` | 初始化函数，返回值供依赖注入 |

### 3.4 method

供 servlet / restrpc 使用：`GET` / `POST` / `PUT` / `DELETE` / `PATCH`。

### 3.5 filters

逗号分隔的函数名（即 filter 函数名）。系统从 filter 集合中查找并绑定。

### 3.6 srpc

定义 **interface + 全局 var**：gos 生成 `XxxStruct` 实现与 `initSrpcClient`，在 init 时把客户端赋给该 var。

1. interface：`type=srpc`，`host` 为 base URL 字面量，或同包内「返回 host 的函数调用」
2. 每个方法：`url` 为相对路径（拼到 `host` 后）；签名 `(ctx, req) (resp, error)`
3. 必须声明 `var Xxx InterfaceType`，否则不会生成注入

```go
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

var BookSvc BookClient
```

业务侧直接调用（无需手写 HTTP）：

```go
resp, err := BookSvc.GetBook(ctx, &GetBookRequest{Id: 1})
```

### 3.7 restrpc

定义 **interface + 全局 var**：gos 生成 `XxxStruct` 实现与 `initRestrpcClient`。

与 srpc 的差异：

1. 使用 HTTP method（默认 `POST`；可写 `method=GET|POST|PUT|DELETE|PATCH`）
2. 成功：HTTP 2xx，响应体直接反序列化为返回对象
3. 失败：HTTP 非 2xx，将响应体反序列化为 `*RestError` 并作为 `error` 返回；网络/编解码等本地失败则返回本地 `error`
4. `GET` / `HEAD`：参数编码为 query；其他 method：JSON body

```go
func GetOrderHost() string {
	return "http://order.internal"
}

// @gos type=restrpc; host=GetOrderHost()
type OrderClient interface {
	// @gos url="/order/get"; method=GET
	GetOrder(ctx context.Context, req *GetOrderRequest) (*GetOrderResponse, error)

	// @gos url="/order/create"
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
}

var OrderSvc OrderClient
```

```go
resp, err := OrderSvc.GetOrder(ctx, &GetOrderRequest{Id: 1})
if err != nil {
	if restErr, ok := err.(*RestError); ok {
		// HTTP 失败：restErr.StatusCode / Message / Raw
		_ = restErr
	}
	// 否则为本地错误（marshal / network 等）
	return err
}
```

### 3.8 FromEntity 自动生成

schema 上添加 `@gos entity=EntityName` 时，在当前包 `schema.gen.go` 中生成：

1. `FromEntity(e *EntityName)`
2. `FromEntitys(entitys []*EntityName)`（当前 schema 的数组类型）

字段匹配规则：

- 遍历 entity（支持嵌套，最多向内 5 层），按 **同名字段** 赋值
- 嵌套匹配也会生成，例如 schema 的 `ID` ← `e.BaseInfo.ID`
- 找不到同名字段时编译器 WARNING；可在 schema 上实现 `FormatEntity()`，生成器会在 `FromEntity` 内调用以补充自定义逻辑

---

## 4. 演示与讲解提纲（可选）

用于对外介绍或内部培训时的推进顺序：

1. 后端编写 → 前端显示（胶水代码与前端生成）
2. 引入数据库 → entity / DAL 自动生成
3. 前端列表展示
4. 补充：注入、URL、数据库、filter、结构体与辅助函数
