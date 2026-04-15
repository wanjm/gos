本篇介绍,api接口的前后端联动
gos 根据api接口,自动生成前端http调用interface. 自动匹配函数,url,出参,入参, 自动添加pagecontroller支持;前端用户直接调用函数就可以了
一个接口开发
1. entity介绍,自生成entity到schema的结构体



前面我们已经介绍了gos在server注解, 数据库代码生成方面的功能, 今天我们将继续将能力拓展到前端;向大家展示一下, 当我们定义了一个http api  hello 服务, 运行gos后,前端将自动生成一个hello函数, 前端业务模块直接调用该函数将 完成 结构体序列化, http调用, 返回值反序列等一系列过程,直接使用结果的神奇能力;


服务端定义如下:
```
// @gos url="/hello"; title="hello函数标题"
func (s *SimpleBiz) SayHello(ctx context.Context, req *HelloRequest) (HelloResponse, error) {
	...
}
```
前端的业务的同学可以直接调用 `res=await simpleBizApi.sayHello(req)`; 就可以通过完成http调用的过程,得到返回结果,整个连接过程全部代码全部自动生成;

我们整个的分解过程如下:
1. gos 生成前端业务函数interface,和request 和 response结构体;
2. dart build_runner 自动根据interface, 自动生成代码http调用的代码逻辑,此步依赖component_generator
3. http的代码的库由component_set提供支撑;


## gos生成服务关联信息,包括api定义和结构体定义
### api定义 
1. go项目中的project.public.toml中,配上FlutterPath="../gos_client_demo/lib/data/http",将自动生成FLUTTER的HTTP连接代码;
```
@DataInterface()
abstract class SimpleBiz {
  /// hello函数标题
  static const String sayHelloUrl = "/hello";
  /// hello函数标题
  @ReqConfig(sayHelloUrl)
  Future<RespData<HelloResponse?>> sayHello(HelloRequest data);
}

```

1. 此处我们可以看到, 生成了函数定义, 入参和出参引用
2. HTTP的URL定义;
3. 函数的标题注释;

### 结构体定义
```
// ["name"]
class HelloRequest extends JSONParameter {
 /// 招呼名字
 String name;

  HelloRequest({
  this.name = "",

  });

  factory HelloRequest.fromJson(Map<String, dynamic> json) {
    return HelloRequest(
   name: json['name'] as String? ?? "",
    );
  }

  @override
  Map<String, dynamic> toJson() {
    return {
      "name": name,
    };
  }
}

// ["message"]
class HelloResponse extends JSONParameter {
 /// 招呼内容
 String message;

  HelloResponse({
  this.message = "",

  });

  factory HelloResponse.fromJson(Map<String, dynamic> json) {
    return HelloResponse(
   message: json['message'] as String? ?? "",
    );
  }

  @override
  Map<String, dynamic> toJson() {
    return {
      "message": message,
    };
  }
}
```
1. 结构体定义;
2. 变量注释;
3. toJson,fromJson 序列化函数;


### 接下来生成前端报文处理真实代码 
运行 `dart run build_runner build` (component_generator)
前端将继续自动生成:
```
class SimpleBizApi extends BaseMethod implements SimpleBiz {
  SimpleBizApi({required super.client});

  @override
  Future<RespData<HelloResponse?>> sayHello(HelloRequest data) => getData(
    data: data,

    url: SimpleBiz.sayHelloUrl,
    buffer:
        bufferMap[SimpleBiz.sayHelloUrl] as ClassBuffer<int, HelloResponse>?,

    encodeDataFunction: (RespData resp) {
      resp.obj = HelloResponse.fromJson(resp.res);
    },
  );
}

var simpleBizApi = SimpleBizApi(client: client);
```

1. SimpleBiz 即前面GOS自动生成的代码
2. BaseMthod 是来源于component_set的底层HTTP支撑;
3. 从生成的代码中,我们还能看到有BUFFER的能力,后续再说;


### 有AI,还要这些吗?
1. 本架构自动生成代码,代码结构稳定,且完全符合预期;
2. 代码自动生成,节省token,且可以反复生成;
3. 代码自动生成,结构升级,重新运行,即可全量秒级生成代码,让重构不再有压力;

### Ai能干嘛?
配上skill, 你主要告诉ai,我要什么XX业务,并在前端获取数据并显示;
1. Ai可以自动后端接口定义和业务代码;
2. ai可以自动运行gos生成前端代码
3. ai可以自动前端实体代码
4. ai可以自动完成函数调用,完成前端业务;
5. 前后端联动, Ai直接知道后端和前端的对应关系,减少人机沟通成本;
6. 有了这个框架,让你可以更好的驾驭AI (harness engineering 从这里起步)
7. ai可以帮助我们写框架, 我们在用框架生成代码. 如本文中的gos, component_generator,component_set都是笔者要求ai编写的;

### 此致前后端联动已经打通, 后续将继续在前后端生成自动的代码,让代码自动化覆盖更观;
o



Go Web 开发：API 与前端一键联动，告别接口对接成本
我将生成适配知乎风格的简短亮点摘要，放在文章开头，突出核心价值、适配专栏封面展示，同时贴合全文调性，不冗余、抓重点：

---
Go Web 开发提速 6（gos）：API 前后端联动生成 —— 一键打通全栈，告别接口对接成本
本文亮点
✅ 全栈自动化闭环：Go 后端接口注解定义 → 前端调用代码一键生成，无需手写 HTTP、序列化逻辑
✅ AI + 框架双提效：框架保稳定、AI 做创意，重构秒级完成，零 Token 消耗、零沟通成本
✅ 依赖清晰可复用：基于 component_generator 实现 build_runner 生成，component_set 提供底层支撑，Demo 开源可直接参考
本系列持续更新，前五篇回顾：
1. [Go Web 开发提速：基于 Spring 式注释方案，用 gos 自动生成运行代码](sslocal://flow/file_open?url=https%3A%2F%2Fzhuanlan.zhihu.com%2Fp%2F1937905040842004437&flow_extra=eyJsaW5rX3R5cGUiOiJjb2RlX2ludGVycHJldGVyIn0=)
2. [Go Web 开发提速 (gos)：Servlet 注解与参数解析全指南 —— 从定义到落地](sslocal://flow/file_open?url=https%3A%2F%2Fzhuanlan.zhihu.com%2Fp%2F1937994788919019061&flow_extra=eyJsaW5rX3R5cGUiOiJjb2RlX2ludGVycHJldGVyIn0=)
3. [Go Web 开发提速 3（gos）：Filter 实战与变量注入 —— 通用逻辑复用与依赖解耦](sslocal://flow/file_open?url=https%3A%2F%2Fzhuanlan.zhihu.com%2Fp%2F1942992392115446822&flow_extra=eyJsaW5rX3R5cGUiOiJjb2RlX2ludGVycHJldGVyIn0=)
4. [Go Web 开发提速 4（gos）：自动生成代码实战解析，破除 Spring 注入误解](sslocal://flow/file_open?url=https%3A%2F%2Fzhuanlan.zhihu.com%2Fp%2F1952837653339828295&flow_extra=eyJsaW5rX3R5cGUiOiJjb2RlX2ludGVycHJldGVyIn0=)
5. [Go Web 开发提速 5（gos）：数据库代码全自动生成 —— 多库统一 + 零硬编码 + 极致复用](sslocal://flow/file_open?url=https%3A%2F%2Fzhuanlan.zhihu.com%2Fp%2F1994878858147686183&flow_extra=eyJsaW5rX3R5cGUiOiJjb2RlX2ludGVycHJldGVyIn0=)
在前五篇中，我们已经完整落地了 gos 在服务端的全链路能力：从 Spring 风格注解驱动、Servlet 接口定义、Filter 拦截与复用、依赖自动注入，到数据库层代码全自动生成，基本实现了服务端 “少写甚至不写重复代码”。
本篇，我们将 gos 的能力正式延伸到前端，实现服务端 API 定义 → 前端调用代码一键生成的全栈联动。
只需要在 Go 中定义一个 HTTP API 接口，运行 gos 后，前端就能自动生成对应的调用函数，前端业务直接调用即可完成结构体序列化、HTTP 请求、返回值反序列化等全套流程，直接使用结构化结果。

---
一、效果先看：一行调用，打通前后端
服务端定义
// @gos url="/hello"; title="hello函数标题"
func (s *SimpleBiz) SayHello(ctx context.Context, req *HelloRequest) (HelloResponse, error) {
	// ...业务逻辑
}
前端业务直接调用
final res = await simpleBizApi.sayHello(req);
整个网络链路、数据转换、异常处理全部自动生成，无需手写任何 HTTP 相关代码。

---
二、自动化流程拆解
整体流程分为三步：
1. gos 生成前端接口抽象、Request/Response 结构体；
2. dart build_runner 基于 component_generator 自动生成真实 HTTP 调用逻辑；
3. 底层网络能力由 component_set 提供支撑。

---
三、gos 生成服务元信息：API 定义 + 结构体定义
1. 配置前端输出路径
在 Go 项目的 project.public.toml 中配置：
FlutterPath="../gos_client_demo/lib/data/http"
gos 会自动将前端代码生成到该目录。
2. API 接口自动生成
@DataInterface()
abstract class SimpleBiz {
  /// hello函数标题
  static const String sayHelloUrl = "/hello";

  /// hello函数标题
  @ReqConfig(sayHelloUrl)
  Future<RespData<HelloResponse?>> sayHello(HelloRequest data);
}
包含：
- 接口 URL 常量定义
- 完整注释同步
- 严格对齐的入参、出参类型
- 标准异步方法签名
3. 结构体自动生成
// ["name"]
class HelloRequest extends JSONParameter {
  /// 招呼名字
  String name;

  HelloRequest({
    this.name = "",
  });

  factory HelloRequest.fromJson(Map<String, dynamic> json) {
    return HelloRequest(
      name: json['name'] as String? ?? "",
    );
  }

  @override
  Map<String, dynamic> toJson() {
    return {
      "name": name,
    };
  }
}

// ["message"]
class HelloResponse extends JSONParameter {
  /// 招呼内容
  String message;

  HelloResponse({
    this.message = "",
  });

  factory HelloResponse.fromJson(Map<String, dynamic> json) {
    return HelloResponse(
      message: json['message'] as String? ?? "",
    );
  }

  @override
  Map<String, dynamic> toJson() {
    return {
      "message": message,
    };
  }
}
包含：
- 完整字段结构与默认值
- 字段注释同步
- toJson / fromJson 自动实现
- 统一基类，适配全局报文规范

---
四、前端执行 build_runner 生成真实调用代码
执行命令：
dart run build_runner build
注：dart build_runner 的生成逻辑基于 component_generator 实现。
自动生成 HTTP 实现类：
class SimpleBizApi extends BaseMethod implements SimpleBiz {
  SimpleBizApi({required super.client});

  @override
  Future<RespData<HelloResponse?>> sayHello(HelloRequest data) => getData(
    data: data,
    url: SimpleBiz.sayHelloUrl,
    buffer: bufferMap[SimpleBiz.sayHelloUrl] as ClassBuffer<int, HelloResponse>?,
    encodeDataFunction: (RespData resp) {
      resp.obj = HelloResponse.fromJson(resp.res);
    },
  );
}

var simpleBizApi = SimpleBizApi(client: client);
- BaseMethod 来自 component_set 提供的底层 HTTP 支撑
- 自动完成请求、序列化、反序列化
- 内置 buffer 扩展，支持缓存扩展
至此，前后端联动代码全部生成完毕：
- 前后端结构永远同步
- 注释自动携带，IDE 智能提示
- URL 统一管理，无硬编码
- 不再依赖 Swagger 做前后端沟通

---
五、有 AI，还需要这套架构吗？
很多人会问：现在 AI 这么强，手写代码都少了，还需要这种代码生成框架吗？
答案是：非常需要，而且是 AI 时代更需要。
1. 本架构生成的代码结构稳定、风格统一、完全符合预期，不会出现 AI 随机 “脑洞” 写法；
2. 自动生成不消耗 Token，且可以随时重新生成，成本极低；
3. 架构升级、结构调整后，重新运行即可秒级全量更新，让重构几乎无压力。
AI 擅长 “创作”，而框架擅长 “规范、稳定、可复现”。
AI + 稳定代码生成框架，才是真正的提效组合。

---
六、AI 能在这套体系里干什么？
给 AI 配上工程化 skill 后，你只需要告诉 AI：
我要实现 XX 业务，前端获取数据并展示。
AI 就可以完成一整套全栈闭环：
1. AI 自动完成后端接口定义与业务逻辑代码；
2. AI 自动调用 gos，生成前端接口与结构体；
3. AI 自动生成前端实体、页面组件；
4. AI 自动完成接口调用与业务逻辑拼装；
5. 前后端联动结构明确，AI 天然知道前后端对应关系，大幅减少人机沟通成本；
6. 这套框架让你真正能驾驭 AI，harness engineering 从这里起步；
7. AI 可以帮我们写框架，我们再用框架规模化生成代码。
本文中的 gos、component_generator、component_set 本身，就是笔者通过 AI 辅助设计与编写的。
简单说：
AI 负责创意与业务，框架负责规范与稳定，各司其职。

---
七、示例代码地址
本文涉及的 DEMO 可在以下仓库查看：
- component_generator：https://github.com/wanjm/component_generator
- component_set（demo1 分支）：https://github.com/wanjm/component_set

---
八、总结与后续
至此，gos 前后端联动已经完整打通。
后续我们将继续扩展代码自动化能力，让自动生成覆盖更多场景、更深链路，真正实现 “定义即代码、结构即产品” 的极致开发体验。
