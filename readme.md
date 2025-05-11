# 准备
1. 我需要设计一个扫描go语言代码，解析出其中关于servlet的定义部分，然后自动跟gin的对接代码，完成端口监听并根据url调用servlet的功能。
2. 当定义一个函数 如下格式时，自动生成/hello的servet代码；并将请求解析为schema.HelloRequest类型，返回schema.HelloResponse类型。
```go
type Hello struct {

}
// @goservlet url="/hello";
func (hello *Hello) SayHello(ctx context.Context, req *schema.HelloRequest) (res schema.HelloResponse, err basic.Error) {
```
# 设计
1. 我们需要定义一个Project结构体，用于存储项目的信息。
2. 我们需要定义一个Package结构体，用于存储package的信息。
3. 我们需要定义一个Struct结构体，用于存储struct的信息。
4. 我们需要定义一个Function结构体，用于存储function的信息。

## 代码
创建astinfo目录，用于存放设计代码

## Project
1. 代码存放存放在astinfo/project.go中
2. Project结构体用于存储项目的信息。
3. Project结构体包含以下字段：
    - Name: 项目名称
    - Module: 项目的module名称
    - Packages: 项目中包含的package map key为package的全路径；value为Package结构体

## Package
1. 代码存放存放在astinfo/package.go中
2. Package结构体用于存储package的信息。
3. Package结构体包含以下字段：
    - Name: package名称
    - Structs: package中包含的struct map key为struct的名称；value为Struct结构体

## Struct
1. 代码存放存放在astinfo/struct.go中
2. Struct结构体用于存储struct的信息。
3. Struct结构体包含以下字段：
    - Name: struct名称

