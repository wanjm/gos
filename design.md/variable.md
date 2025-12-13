## map的解析与生成；
1. 添加MapType类，保证了正常解析；
2. 并且添加到isRawType中，跳过对依赖注入时来源的检查；
3. 生成代码时，特别处理该类型，调用Type.GenConstructCode，生成初始化方法；

## channel的解析与生成；（）
1. 与map类似
2. 暂时不支持size的写法；
