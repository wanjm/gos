本篇介绍：API 接口的 TypeScript 前后端联动
gos 根据 api 接口，自动生成前端 TS HTTP 调用代码：自动匹配函数、url、出参、入参；前端业务直接 `await xxxApi.method(req)` 即可。

前面我们已经介绍了 gos 在 server 注解、数据库代码生成方面的功能，第 6 篇把能力拓展到了 Flutter。今天继续把同一套后端定义，延伸到 **TypeScript**（Web / 小程序等）：定义一个 HTTP API，运行 gos 后，TS 侧自动生成调用类与类型，业务直接拿到结构化结果。


服务端定义如下：
```
// @gos url="/hello"; title="hello函数标题"
func (s *SimpleBiz) SayHello(ctx context.Context, req *HelloRequest) (HelloResponse, error) {
	...
}
```
前端业务同学可以直接调用：
`const res = await simpleBizApi.sayHello(req);`
即可完成 HTTP 调用，整个连接过程全部自动生成。

我们整个的分解过程如下：
1. gos 生成 TS 的 API 类、`schema.gen.ts` 类型定义（以及可选的 `enum.gen.ts`）；
2. 首次生成时写入可手改的 `platform.ts`（选择 WxClient / PcClient、配置 baseUrl / header）；
3. 底层 HTTP 由 npm 包 `component_set/http` 提供（`Business`、`RespClass`、`WxClient` / `PcClient`）。

与第 6 篇 Flutter 路径的差异：TS **不需要**再跑 `build_runner`——gos 一次写出可直接调用的实现类。


## gos 生成服务关联信息：API 定义 + 结构体定义
### 配置
在 Go 项目的 `project.public.toml` 中配置：
```
TsPath="../tboss/src/data/http"
```
（yxt_data 示例；也可指向小程序/Web 工程目录，如 `../xxx/src/data/http`）

### API 定义（network.gen.ts）
```ts
export class SimpleBizApi extends Business {
  /** hello函数标题 */
  static readonly sayHelloUrl = "/hello";
  /** hello函数标题 */
  async sayHello(req: HelloRequest): Promise<RespClass<HelloResponse>> {
    return this.getData<HelloResponse>(SimpleBizApi.sayHelloUrl, req);
  }
}

export const simpleBizApi = new SimpleBizApi(client);
```

1. 生成了 API 类与方法；
2. URL 静态常量；
3. title 注释同步；
4. 全局实例可直接 import 使用。

### 结构体定义（schema.gen.ts）
```ts
export interface HelloRequest {
  /** 招呼名字 */
  name: string;
}

export interface HelloResponse {
  /** 招呼内容 */
  message: string;
}
```

1. interface 定义；
2. 字段注释同步；
3. 嵌套类型、数组、number/boolean/string 自动映射。


### platform.ts（只写一次，可自由修改）
gos 若发现目录下尚无 `platform.ts`，会写入模板；之后**不会覆盖**，方便业务改 baseUrl、Token、切平台：

```ts
// Written once by gos; edit freely.
import { WxClient } from "component_set/http";
// import { PcClient } from "component_set/http";

export const client = new WxClient();
// export const client = new PcClient();
```

真实项目里常会扩展 Client（设置 `ApiBaseUrl`、注入 token header），例如小程序侧：

```ts
import { WxClient as PlatformClient } from "component_set/http";

class Client extends PlatformClient {
  constructor() {
    super();
    this.ApiBaseUrl = "https://api.example.com";
  }
  // getHeaders 中附加 token ...
}

export const client = new Client();
```


### 有 AI，还要这些吗？
1. 本架构自动生成代码，结构稳定、完全符合预期；
2. 代码自动生成，节省 token，且可反复生成；
3. 结构升级后重新运行 gos，即可全量秒级更新，重构压力小。

### AI 能干嘛？
配上 skill，你主要告诉 AI：我要什么 XX 业务，并在 TS 前端获取数据并显示；
1. AI 可以自动后端接口定义和业务代码；
2. AI 可以自动运行 gos 生成前端 TS 代码；
3. AI 可以自动完成函数调用与页面拼装；
4. 前后端联动，AI 直接知道对应关系，减少人机沟通成本；
5. harness engineering：框架保规范，AI 做业务。


Go Web 开发：TypeScript HTTP Client 一键联动
我将生成适配知乎风格的简短亮点摘要，放在文章开头：

---
Go Web 开发提速 7（gos）：TypeScript HTTP Client 自动生成 —— Web / 小程序同样一键联动
本文亮点
✅ 同一份 Go API 注解：配置 TsPath 后，一键生成 network / schema（及 enum）TS 代码
✅ 开箱即用：继承 component_set 的 Business，直接 await xxxApi.method(req)
✅ 平台可切换：platform.ts 只写一次，WxClient / PcClient 自由切换，gos 不覆盖手改

本系列持续更新，前六篇回顾：
1. [Go Web 开发提速：基于 Spring 式注释方案，用 gos 自动生成运行代码](https://zhuanlan.zhihu.com/p/1937905040842004437)
2. [Go Web 开发提速 (gos)：Servlet 注解与参数解析全指南 —— 从定义到落地](https://zhuanlan.zhihu.com/p/1937994788919019061)
3. [Go Web 开发提速 3（gos）：Filter 实战与变量注入 —— 通用逻辑复用与依赖解耦](https://zhuanlan.zhihu.com/p/1942992392115446822)
4. [Go Web 开发提速 4（gos）：自动生成代码实战解析，破除 Spring 注入误解](https://zhuanlan.zhihu.com/p/1952837653339828295)
5. [Go Web 开发提速 5（gos）：数据库代码全自动生成 —— 多库统一 + 零硬编码 + 极致复用](https://zhuanlan.zhihu.com/p/1994878858147686183)
6. [Go Web 开发提速 6（gos）：API 前后端联动生成 —— 一键打通全栈，告别接口对接成本](https://zhuanlan.zhihu.com/p/2027513197754589261)

第 6 篇把 gos 延伸到了 Flutter：Go 定义 API → Dart 调用代码一键生成。
本篇平行能力落在 **TypeScript**：同一套 servlet 注解，配置 `TsPath` 后，Web / 小程序前端同样获得强类型 HTTP Client。

---
一、效果先看：一行调用，打通前后端
服务端定义
```go
// @gos url="/hello"; title="hello函数标题"
func (s *SimpleBiz) SayHello(ctx context.Context, req *HelloRequest) (HelloResponse, error) {
	// ...业务逻辑
}
```
前端业务直接调用
```ts
const res = await simpleBizApi.sayHello(req);
// res: RespClass<HelloResponse>
```
网络请求、类型约束、统一响应封装全部自动生成，无需手写 fetch / axios 样板代码。

真实项目示例（图书列表）：
```ts
const res = await bookBizApi.bookList({ pageNum: 1, pageSize: 20 });
```

---
二、自动化流程拆解
整体比 Flutter 路径更「一步到位」：
1. gos 扫描 Go servlet，收集 API 与入参/出参结构体；
2. 写出 `schema.gen.ts`、`network.gen.ts`（有枚举则另写 `enum.gen.ts`）；
3. 若缺少 `platform.ts`，写入可编辑模板；底层能力来自 `component_set/http`。

无需 dart 的 build_runner 第二阶段——生成物本身就是可调用的实现类。

---
三、配置 TsPath
在 Go 项目 `project.public.toml`：
```toml
[Generation]
FlutterPath="../boss/lib/data/http"
TsPath="../tboss/src/data/http"
```
`TsPath` 为空则跳过 TS 生成；与 Flutter 可同时开启，一份后端定义服务两端。

---
四、gos 生成了哪些文件？
| 文件 | 是否每次覆盖 | 作用 |
|------|--------------|------|
| `schema.gen.ts` | 是 | Request / Response 等 interface |
| `network.gen.ts` | 是 | `*Api` 类 + 全局实例 |
| `enum.gen.ts` | 是（有 enum 时） | 枚举与文案 helper |
| `platform.ts` | 仅缺失时写入 | 选择平台 Client、baseUrl、header |

### 1. schema.gen.ts
```ts
export interface HelloRequest {
  /** 招呼名字 */
  name: string;
}

export interface HelloResponse {
  /** 招呼内容 */
  message: string;
}
```
- Go 基础类型映射为 `string` / `number` / `boolean`
- 结构体 → 同名 interface；数组 → `T[]`
- 字段注释同步；`json:"-"` 字段跳过

### 2. network.gen.ts
```ts
import { Business, RespClass } from "component_set/http";
import { client } from "./platform";
import type { HelloRequest, HelloResponse } from "./schema.gen";

export class SimpleBizApi extends Business {
  /** hello函数标题 */
  static readonly sayHelloUrl = "/hello";

  /** hello函数标题 */
  async sayHello(req: HelloRequest): Promise<RespClass<HelloResponse>> {
    return this.getData<HelloResponse>(SimpleBizApi.sayHelloUrl, req);
  }
}

export const simpleBizApi = new SimpleBizApi(client);
```
包含：
- 类名：servlet 结构体名 + 必要时补 `Api` 后缀
- URL：结构体 url 前缀与方法 url 拼接
- `getData`：委托给 `Business` → platform `client`
- 导出单例，业务侧直接 import

### 3. platform.ts 与 component_set
HTTP 运行时在 npm 包 `component_set`：
- `Business`：生成类的基类，持有 client 并代理 `getData`
- `RespClass<T>`：统一响应包装（与后端 `{code,msg,obj}` 风格对齐）
- `WxClient` / `PcClient`：小程序 / PC 平台实现

切换平台只需改 `platform.ts` 的 import，**不必改** `network.gen.ts`。

---
五、与第 6 篇 Flutter 对照
| 项 | Flutter（第 6 篇） | TypeScript（本篇） |
|----|-------------------|-------------------|
| 配置项 | `FlutterPath` | `TsPath` |
| 接口抽象 | `@DataInterface` abstract + build_runner | 直接生成 `*Api extends Business` |
| 结构体 | class + fromJson/toJson | `export interface` |
| 二次生成 | 需要 `dart run build_runner` | 不需要 |
| 运行时库 | component_set（Dart） | component_set/http（npm） |

后端注解完全相同；差别只在 Generation 输出路径与目标语言。

---
六、有 AI，还需要这套架构吗？
1. 生成代码结构稳定、风格统一，避免 AI 随机写法；
2. 自动生成不消耗 Token，可随时重跑；
3. API 增删改后重新 gos，前后端类型一起刷新，重构成本低。

AI 负责业务与页面，框架负责契约与稳定——与第 6 篇同一套 harness engineering 思路。

---
七、AI 能在这套体系里干什么？
告诉 AI：实现 XX 业务，TS 前端拉列表并展示。
1. AI 完成后端 servlet / schema；
2. AI 运行 gos，生成 `network.gen.ts` / `schema.gen.ts`；
3. AI 在页面 `import { courseApi } from '...'` 并调用；
4. 前后端类型同源，减少「字段对不上」的沟通成本。

---
八、示例与依赖
- gos：https://github.com/wanjm/gos
- component_set（HTTP 运行时）：业务侧依赖 npm `component_set`
- 配置示例：yxt_data 的 `TsPath="../tboss/src/data/http"`
- 生成形态参考：小程序工程中的 `network.gen.ts` + `platform.ts`

---
九、总结与后续
至此，gos 在 Flutter 之外，打通了 **Go API → TypeScript HTTP Client** 的全链路。

同一份注解，同时服务 Boss（Dart）与 Web/小程序（TS）；后续第 8 篇将介绍 `@gos enum`，让状态码与展示文案也在多端自动对齐。
