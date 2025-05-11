package astinfo

type Project struct {
	Name     string
	Module   string
	Packages map[string]*Package // key为package全路径
}

func NewProject() *Project {
	return &Project{
		Packages: make(map[string]*Package),
	}
}