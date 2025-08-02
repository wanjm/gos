# project 产生代码
1. server注册；
2. 全局变量生成；
3. 全局变量初始化；
4. 全局变量的使用；
5. 依赖注入，及按照依赖关系保证代码生成顺序；
6. 路由注册函数
7. 


# 解析import遇到的意外；
## 原以为每个import都在go.mod中，但是系统的package是个例外，其在GOROOT中；
## 原以为可以跳过系统package的解析，但是，还是有很多结构体都是在GOROOT中的。暂时还不解析，看看还会遇到什么问题
## 缓存目录大小写变动github.com/BurntSushi/toml => github.com/!burnt!sushi/toml
## 原以为不会存在两个pkg（project）其中一个是另一个的子目录；
如下： 这是两个不同的project，他们有自己的版本，v2不是第一个的子目录；
github.com/pelletier/go-toml v1.9.5
github.com/pelletier/go-toml/v2 v2.2.2 
## 原以为一个目录只有会一个pkg，实际可能有多个
1. ***
2. ***_test
3. 其他如main，但是文件中有 "//go:build ignore"注释；
   
## 原以为每个go的import都对应go文件；
1. 有//go:build ignore

## 奇怪经历
1. 遇到了系统调用go的库，但是不存在的情况；
