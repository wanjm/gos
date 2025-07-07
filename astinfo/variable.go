package astinfo

// Field是源码中定义的变量；
// Variable时是生成的代码中定义的用来使用的变量；
type Variable struct{}

// 当需要一个变量值时如下几个场景；
// 该变量类型在全局函数存在，则从全局变量获取，直接返回变量名即可
// reciver.function creator!=nil, receiverPrex!=""
// schema.struct
// schema.function  creator!=nil, receiverPrefix==""
// 返回值无\n
func (v *Variable) Generate(goGenerated *GoGenerated) error {
	return nil
}
