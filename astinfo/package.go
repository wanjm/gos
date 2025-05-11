package astinfo

// Package 表示一个Go包的结构信息
type Package struct {
	Name    string             // 包名称（与目录名一致）
	Module  string             // 包的全路径（如"github.com/user/project/pkg"）
	Structs map[string]*Struct // 包内定义的结构体集合
}

// NewPackage 创建并初始化一个新的Package实例
func NewPackage() *Package {
	return &Package{
		Structs: make(map[string]*Struct),
	}
}
