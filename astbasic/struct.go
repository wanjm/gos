package astbasic

import (
	"path"
	"path/filepath"
)

type PkgBasic struct {
	Name     string
	ModPath  string
	FilePath string // 包所在目录的绝对路径
}

func (p *PkgBasic) NewPkgBasic(name, pathValue string) *PkgBasic {
	var modPath string
	if !filepath.IsAbs(pathValue) {
		modPath = path.Join(p.ModPath, pathValue)
		pathValue = filepath.Join(p.FilePath, pathValue)
	}
	return &PkgBasic{
		Name:     name,
		FilePath: pathValue,
		ModPath:  modPath,
	}
}

func (p *PkgBasic) NewFile(fileName string) *GenedFile {
	file := CreateGenedFile(fileName)
	file.Pkg = p
	return file
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
