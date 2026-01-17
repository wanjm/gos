# flutter gen
1. 本项目会自动解析golang结构体和函数，生成Struct，Method和Function;
2. 对于一个定义为如下的golang servlet请求；
```
// @gos type=servlet; url="/example"
type Hello struct {
	BizHello *BizHello
}

// @gos url="/hello"; title="hello"
func (hello *Hello) SayHello(ctx context.Context, req *HelloRequest) (HelloResponse, error) {
	return HelloResponse{
		Greeting: "hello "+req.Name,
	}
}

type HelloRequest struct {
	Name string `json:"name"`
}
type HelloResponse struct {
	Greeting string `json:"greetintg"`
}
```
3. 希望自动生成flutter调用代码；
```
class HelloRequest extends JSONParameter {
  final String name;
  final int loginType = 0;

  HelloRequest(this.name);

  @override
  Map<String, dynamic> toJson() {
    return {
      "name": name,
    };
  }
}

class HelloResponse {
  final String name;
  final int loginType = 0;

  HelloResponse(this.name);

  @override
  Map<String, dynamic> toJson() {
    return {
      "name": name,
    };
  }
  factory HelloResponse.fromJson(Map<String, dynamic> json) {
    return HelloResponse(json['name'] ?? '');
  }
}

abstract class Network {
  Future<RespData<HelloResponse?>> sayHello(HelloRequest data);
}

class NetworkImpl extends BaseMethod implements Network {
  NetworkImpl({required MyClient client}) : super(client: client);
  @override
  Future<RespData<HelloResponse?>> sayHello(HelloRequest data) => getData(
    data: data,
    url: "/hello",
    encodeDataFunction: (RespData resp) {
      resp.obj = HelloResponse.fromJson(resp.res);
    },
  );
}

var network = NetworkImpl(client: client);
```
5. 其中函数的url和函数名来源于golang的定义，且符合dart的规则；
6. 结构体定义来源于golang的定义，且符合dart的规则；
7. 对于结构体需要定义tojson和fromJson方法， 内容为dart 的json_serializable 生成的内容；
8. 对于对象中为数组的，如果服务端返回为null，则初始化为空数组；


