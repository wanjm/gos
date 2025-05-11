package astinfo

type Package struct {
	Name    string
	Module  string // 存储package的全路径（如"github.com/user/project/pkg"）
	Structs map[string]*Struct
}

func NewPackage() *Package {
	return &Package{
		Structs: make(map[string]*Struct),
	}
}