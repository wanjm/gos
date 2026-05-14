# key
1. url  //定义服务路径
2. title //定义服务标题；
3. autogen // 类自动生成对象，供注入；
4. group // 服务所属的群组，修饰struct，和filter；
5. tblName // 修饰entity的tableName
6. dbVariable // 修饰entity的dbVariable， dal中链接驱动的接入；
7. host //prpc，srpc客户端，提供获取host的地址函数；
8. type // 类型；
9.  method //方法；
10. fillters; //过滤器函数名；

## type
表示struct，method，function的类型；
1. servlet; （类是servlet）；
2. prpc；
    - struct prpc服务端
    - interface 表示prpc客户端
3. srpc； （client是srpc客户端）
4. filter？ 表示是filter函数；
   

## method
供servlet使用；
1. GET
2. POST

## filters
逗号分隔的函数名。直接是filter函数的名字。 系统会自动从filter中去找；