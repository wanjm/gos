package astbasic

type PkgBasic struct {
	Name     string
	ModPath  string
	FilePath string // 包所在目录的绝对路径
}

func (p *PkgBasic) IsSame(antother *PkgBasic) bool {
	return p == antother || p.ModPath == antother.ModPath
}

func SimplePackage(module, name string) *PkgBasic {
	return &PkgBasic{
		ModPath: module,
		Name:    name,
	}
}
