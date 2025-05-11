# 准备
1. 我需要设计一个扫描go语言代码，解析出其中关于servlet的定义部分，然后自动跟gin的对接代码，完成端口监听并根据url调用servlet的功能。
2. 当定义一个函数 如下格式时，自动生成/hello的servet代码；并将请求解析为schema.HelloRequest类型，返回schema.HelloResponse类型。
```go
type Hello struct {

}
// @goservlet url="/hello";
func (hello *Hello) SayHello(ctx context.Context, req *schema.HelloRequest) (res schema.HelloResponse, err basic.Error) {
```
