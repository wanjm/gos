package astinfo

type Package struct {
	Name    string
	Structs map[string]*Struct // key为struct名称
}

func NewPackage() *Package {
	return &Package{
		Structs: make(map[string]*Struct),
	}
}