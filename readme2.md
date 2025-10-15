# 
# 解析说明
## 整体解析说明
```mermaid
flowchart TD
A[解析目标工程的配置文件] --> B[解析源码] --> c[生成源码]
```
## 源码解析过程
```
Project.parse => dirs.foreach(Package.parse)
Package.parse => file.foreach(goFile.parse)
goFile.parse => type.switch{create parser;parser.parse}
type => struct, interface, function, method ,variable
```
```
先从go.mod中解析依赖的progject；
在解析自己这个project。这样保证了自己依赖的package都已经被知道；
但是本工程内部的解析过程，由于是按照目录顺序解析的，所以可能会出现依赖的package还没有被解析的情况。
```
### goSource解析
1. import记录的保存；
### Function 解析
1. 解析comments
2. 解析入参和出参
### Type解析
type解析支持如下几个情况
1. type struct；
2. type interface；
3. type A B
4. type A=B
5. 其中B又分为原始类型，外pkg类型，本pkg类型； type A=[pkg.]B, type A [pkg.]B
解析过程；
分为三种类型解析 struct，interface，alias(解决3，4，5组合的四种情况)，其中alias中的B的解析顺序不固定，为了避免多次扫描，如果B是外包，那么B肯定已经存在（因为通过依赖关系去解析），对于B是本包，但是不存在的情况， 则需要map缓存，等解析完成时再连接起来；
Type需要嵌套的；
### const ,var 解析

## 对象被管理
### 函数管理
函数在go 中分别为function和method；
1. function由package管理；
2. method由struct管理；
3. 这两者都抽象为FunctionManager进行管理
### struct 管理
struct 由package管理

## 对象管理
### FunctionManager
FunctionManager 管理了
1. initiator初始化函数；
2. servlet
3. creator
4. postAction 该函数在servlet执行后执行的函数；

### initiatorManager
1. 管理initor初始化函数的返回值，便于注入时查找；
2. 完成initiator初始化函数的调用，并维护依赖关系；

### ServerManager
负责对配置的每个server进行初始化，管理其中的filter，servlet；并生成最终代码中的server代码。打通filter和servlet的注册环节

### Package管理
1. 项目内部解析出的Packge对象；由解析过程中产生；
2. 项目内部饮用的Package对象；由Import语句引入；







# 产品定义
## 函数定义
1. type 该函数的功能说明rpc,filter,initiator,servlet,restful，creator
2. group 
3. method 该函数对外服务的方法 GET,POST
4. title 该函数的说明
5. url 该函数服务的url
其中type需要细化为
1. function type =〉 filter，initiator
2. method type =〉rpc, servlet，restful，postAction, creator

## struct定义
1. type 该struct的功能说明rpc,servlet,restful
2. group 
2. title 该struct的说明
3. url 该struct服务的url前缀
## rpc client定义；
1. type=prpc,servlet;
2. 此处用到全局VAR扫描，仅仅只是var AClient ClientInterface这一种情况；function,channel 
# 开发技巧
## funtion/method 将自己塞到functionManager中去；
1. function/method是被functionManager管理的，那是由functionManager来管理她，还是她把自己送到functionManager中去呢？
2. 目前由于统一使用createParser(参见goFile.parse)，然后调用了Parser.parse函数，导致让functionManager来管理function/method这个模式不现实， 因为其他代码没有这个逻辑，所以目前采用了function/method将自己送到functionManger的方法来实现；
3. 另外method在没有被解析之前，是不知道自己属于哪个functionMaanager的。所以也导致了他们需要将自己送到functionManager中去；



# 代码生成说明
## 代码业务逻辑执行顺序
排序执行顺序定义的生成代码的执行逻辑，后续代码需要按照该逻辑生成代码；
生成的代码最终对外暴露Prepare和Run函数；
1. Prepare完成代码初始化和注入工作；用于非servlet的功能，如cronjob或者test等；
2. Run则整个业务逻辑，filter，路由注入等；（该函数会主动调用prepare）
```mermaid
flowchart TD 
A[Run说明] --> B[initVariable] --> c[initFilter] --> d[initServlet]
P[Prepare说明] --> P1[initVariable]
```
## initiator返回的管理
1. 由于initiator最后需要全局管理，全局排序；
2. 所以initiator一开始由function管理；
3. 但是最终按照相互依赖关系排序到project中；
4. 对于initiator变量，采用一个map来管理，其key是全局唯一表示某个struct的字符串，其值是返回同类型的数组；
5. 返回值只讨论对象和一级指针；多级指针暂时不考虑；
### initiator函数生成；
1. initVariable按照各个package的initorator函数依赖关系依次调用；
2. 调用顺序首先保证依赖顺序；
3. 没有依赖关系的同级函数按照package顺序排序；
4. 同package中的函数按照函数名字排序；
5. 建立[依赖关系](#initiator依赖关系的建立)时，会生成数组，按照数组生成变量，并调用函数即可；
6. 由于initiator的入参仅可能是其他init的结果，所以入参的注入，仅需要到initmanager中寻找即可；

## Callable的生成
servlet生成支持 servlet，prpc，restful；其区别是：
1. servlet， 请求json对象，返回{code:0,msg:"",obj:{}}
2. prpc 请求为[arg1,arg2],然后{c: r:[err,obj]}
3. restful 请求是json对象，返回用http code表示返回code，报文体对应obj；
由于filter有对应提前返回的能力，所以filter与上面的类型配套使用；
因此需要注册一个方法来filter和servlet的生成；

### 生成规则
1. 每个group一个文件；
2. 其中包含了所有的filter；
3. 所有的路由；包含每个package的每个struct，每个方法注册一个；
### Function调用的生成；
function的调用格式是 pkgName.functionName(var1, var2);
1. 其中pkgName来源于function.goSource.pkg;
2. var1来源与function.params[*]=>field;
3. 由function，协助field初始化variable，再由variable产生代码；
4. 其中还包含了pkg的import；

### Method调用的生成；
method.调用格式为 structVariable.methodName(var1, var2); 目前生成servlet调用的代码，都是模版写死的，没有动态生成；
1. 除了structVariable，其他都跟function调用一样；
2. 所以structVariable该如何处理。需要实际遇到再处理；

### Filter


## Field变量的生成；
Field，可以控制Varialbe生成代码的来源；
## Variable变量的生成；
Variable根据自己的来源，确认生成代码；来源请参考[Field](#field变量的生成)
1. creator: 需要返回creator.genCallCode的代码；（暂时不考虑是method的情况）
2. initiator: 需要返回全部变量的变量名；
3. 调用构造写法，此时涉及到其中的每个变量都需要注入（servlet的struct）；
4. 调用构造的写法，使用default值初始化，servlet的request；
5. 如果是servlet的method做creator的情况，则需要动态注入creator；
### 构造的写法
由Typer提供，添加了wire参数，当时servelt对象时，使用注入，request对象不使用注入；实际使用中；
1. servlet的struct会有注入；
2. request的struct不会有注入；
3. 对于原始类型，不写就是默认值；

### initiator依赖关系的建立
算法
1. foreach package；收集initiator functions；并创建variable map，
2. initiator functions按照name排序；
3. foreach initiator functions；
4. 从variableMap中寻找依赖关系，并记录到自己的parent中；
5. 对于每一个initiator,查看其父节点是否ready；
6. 父节点ready，则自己也ready，此时分配__global_xxx的变量名；
7. 放到最终的variableMap中；
8. 将自己放到readyNode列表；
9. 最终剩余functions长度为0, 则完成；
### 变量注入的来源
1. initiator产生的变量；
2. creator产生出来的；
3. 构造函数初始化出来的；

### project全局变量的管理
1. 所有initiaotr的返回值记录到全局管理中；
2. 依靠该变量建立initiator之间的依赖关系；

### 获取需要注入的变量
1. 同名的变量；
2. 同类型的变量；
3. 调用creator
4. 直接写结构体初始化；
为了便于查找变量，需要设计的数据结构；
1. map[*Struct][]Variable;



## 注册servlet函数
## 注册filter函数

## 主函数
1. 调用initiator函数，由于initiator有相互依赖，需要定义执行顺序；
2. 注册filter函数
3. 注册servlet函数

### 数据库
1. 结构体 entity/mysql/className/table.gen.go;
2. 列 entity(表结构体定义)定义的同一个目录 column.gen.go; C_开头；
3. dal 将entity对应的dal生成于entity（目录名）的同级目录；
4. tblName定一个表名，默认为类名转蛇形；dbVariable定义变量名默认DB；