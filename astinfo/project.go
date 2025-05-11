package astinfo

// Project 表示一个完整的Go项目结构
type Project struct {
	Name     string              // 项目名称（通过go.mod解析）
	Module   string              // 项目module名称
	Path     string              // 项目根目录的绝对路径
	Packages map[string]*Package // 项目包含的包集合
}

func NewProject() *Project {
	return &Project{
		Packages: make(map[string]*Package),
	}
}
