## 说明
这个一个模拟spring的工程，通过在注释中配置特定信息的方案，完成servlet的定义，然后通过go_servlet自动生成胶水代码，组建工程，从而简化开发过程，让大家聚焦于业务，少些重复代码；所哟配置全部通过注释完成，无需额外配置文件，自动生成的代码全部在工程中，可以清晰把握代码
注释支持
1. servlet定义；(由于field的存在所以一种server，只能支持一种类型的servlet，不能混合使用)
2. filter定义；
3. 支持自动注入；（支持按类型注入和变量注入）
4. 自动生成swagger；
## servlet开发说明
servlet 通过在函数体上注释url，函数定义入参，返回值的方式完成开发；
### 案例
已开发一个发送 application/json，并返回json的程序为例说明过程；
1. 向url /example/hello 发送请求
2. 请求为 {"name":"world"};
3. 返回结果为{"code":0,"msg":"","obj":{"greeting":"hello world"}}
### servlet 定义
1. 定义servlet服务类
```
// @goservlet type=servlet url="/example"
type Hello struct {
}
```
1. servlet函数定义
```
// @goservlet url="/hello";
func (hello *Hello) SayHello(ctx context.Context, req *schema.HelloRequest) (res schema.HelloResponse, err basic.Error) {
	res.Greeting = "hello " + req.Name
	return
}
```
1. 参数说明

```
type HelloRequest struct {
	Name string `json:"name"`
}
type HelloResponse struct {
	Greeting string `json:"greeting"`
}
```
4. 运行go_servlet生成胶水代码，就可以完成该服务了