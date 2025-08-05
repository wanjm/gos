1. 创建文件的脚本准备
用shell协议touchfile的脚本 根据文件名数组， 检查其对应的go文件是否存在，不存在则，创建，并向文件中添加packge astifo; 生成initgo.sh
2. 将files变量换成为从命令行读入，有多少参数就创建多少文件,如果没有参数则使用默认值
3. 分别在"interface" "function" "method" "variable" "gosourse"对应的go文件中，添加对应的类，并实现Parse(),方法体为空
4. 将刚才的Parse方法添加返回值error
5. 修改Project::ParseModule方法，自己读取第一行，别解析处mode值,通过用空格split第一行，然后获取第一个元素，作为mode值
6. 在WalkDir中如果是.git, gen目录则跳做过解析,用数组记录文件名，便于将来添加更多的文件



# work
1. 错误日志分级，用参数控制；
2. 跳过gen目录的解析；
3. test环境变量的注入；
4. servlet添加filter参数，参数可以配置；
5. 返回值使用error；
6. 添加http的第三个参数；
7. servlet class添加group配置；
8.  GlobalInjector
9.  解决解析中报的各个错误；
10. 添加prpc的实现；
11. go_servlet的parse能力，自动跳过文件版本不匹配或则会ignore的文件，还有其他build标记；
12. 打印依赖关系树
13. package中不需要保存Structs，仅有servlet类型的struct需要保存；可以用service，servlet来保存，其他都用typer保存；
14. field中的匿名结构体和匿名interface还没有解析；
15. var的多行解析需求；

## 暂时不解析：
```
type Pointer[T any] struct {
        // Mention *T in a field to disallow conversion between Pointer types.
        // See go.dev/issue/56603 for more details.
        // Use *T, not T, to avoid spurious recursive type definition errors.
        _ [0]*T

        _ noCopy
        v unsafe.Pointer
}
此处的类型T会报错；
```

# 完成的工作
1. 解析文件
2. 生成了initiator；
3. 生成路由和fileter的调用；
4. 解析外部package的pkg name；
5. 区分简单解析和复杂解析；
6. field的comment解析；
7. struct构造函数添加default配置；
8.  pack的解析按照依赖关系解析
9.  type AiChatRecords []*AiChatRecord
10. type AiAgentStausResp = EmptyResponse 的解析，导致AiAgentStausResp找不到
11. type等多行解析的需求；  
```
type (
    a b 
    c d
)
// this is doc of type 
type A struct { // this is nothing;
    //this is Doc of Age;
    Age int  //this is comments of Age;
} // this is comments of A

type (
    // this is doc of Age 
    A struct { // this is nothing;
        //this is Doc of Age;
        Age int  //this is comments of Age;
    } 
)// this is comments of A

```