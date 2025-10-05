package astbasic

type PkgBasic struct {
	Name    string
	ModPath string
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
