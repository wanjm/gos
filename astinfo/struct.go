package astinfo

// Struct 表示一个Go结构体的基本信息
type Struct struct {
	Name string // 结构体名称
	Pkg  *Package
	// TODO: 后续添加字段和方法解析
}

// new
func NewStruct(name string, pkg *Package) *Struct {
	return &Struct{
		Name: name,
		Pkg:  pkg,
	}
}
func (v *Struct) Parse() error {
	// 方法体为空
	return nil
}
