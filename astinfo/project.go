package astinfo

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// Project 表示一个Go项目的基本信息
type Project struct {
	Name     string              // 项目名称
	Module   string              // 项目模块名称（从go.mod解析）
	Path     string              // 项目根目录的绝对路径
	Packages map[string]*Package // 项目包含的包集合（key为包全路径）
	Cfg      *Config

	*InitManager
	initFuncs []string
}

func (p *Project) ParseModule() error {
	modPath := filepath.Join(p.Path, "go.mod")
	data, err := os.ReadFile(modPath)
	if err != nil {
		return fmt.Errorf("error reading go.mod: %w", err)
	}

	modfile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return fmt.Errorf("error parsing go.mod: %w", err)
	}
	p.Module = modfile.Module.Mod.Path
	fmt.Printf("Module: %s\n", p.Module)
	return nil
}

func (p *Project) Parse() error {
	if err := p.ParseModule(); err != nil {
		return err
	}
	p.Packages = make(map[string]*Package)
	// project.parse(){
	// 	for eachdir {
	//      package.parse
	//  }
	// }
	// pacakge.parse(){}
	//   for eachfile {
	//   	gosource.parse（）
	//   }
	//}
	// gosource.parse（）{
	//         switch type{
	//             // struct
	//             // interface
	//             // variable
	//             // function
	//             // method
	//     }
	// }
	err := filepath.WalkDir(p.Path, func(path string, d fs.DirEntry, err error) error {
		//path是全路径
		if err != nil {
			return err
		}
		// Skip .git and gen directories
		if d.IsDir() {
			dirName := filepath.Base(path)
			skipdirs := []string{".git", "gen"}
			for _, skipdir := range skipdirs {
				if dirName == skipdir {
					return filepath.SkipDir
				}
			}

			if err := p.ParsePackage(path); err != nil {
				return fmt.Errorf("error parsing package at %s: %w", path, err)
			}
		}
		return nil
	})

	return err
}

// dir是pacakge 所在的全路径
func (p *Project) ParsePackage(dir string) error {
	// 计算相对路径
	relPath, err := filepath.Rel(p.Path, dir)
	if err != nil {
		return err
	}
	// 用mode+相对路径，得到包全路径
	pkgPath := filepath.Join(p.Module, relPath)
	pkg := p.FindPackage(pkgPath)
	pkg.Path = dir
	if err := pkg.Parse(); err != nil {
		return fmt.Errorf("package parse error: %w", err)
	}
	return nil
}

// GetPackage retrieves a package by module path without creation
func (p *Project) GetPackage(module string) *Package {
	return p.Packages[module]
}

// FindPackage finds or creates a package with automatic module path resolution
func (p *Project) FindPackage(module string) *Package {
	if pkg := p.GetPackage(module); pkg != nil {
		return pkg
	}
	newPkg := NewPackage(module)
	p.Packages[module] = newPkg
	return newPkg
}

// GenerateCode 生成项目的代码
func (p *Project) GenerateCode() error {
	p.genInitMain()
	// 遍历所有包
	for _, pkg := range p.Packages {
		_ = pkg
		// 生成包的代码
		// if err := pkg.GenerateCode(); err != nil {
		// 	return fmt.Errorf("error generating code for package %s: %w", pkg.Name, err)
		// }
	}
	p.genProjectCode()
	return nil
}

var GlobalProject *Project

func CreateProject(path string, cfg *Config) *Project {
	GlobalProject = &Project{
		Path: path,
		Cfg:  cfg,
		// Package:      make(map[string]*Package),
		// initiatorMap: make(map[*Struct]*Initiators),
		// servers:      make(map[string]*server),
		// creators: make(map[*Struct]*Initiator),
	}
	// 由于Package中有指向Project的指针，所以RawPackage指向了此处的project，如果返回对象，则出现了两个Project，一个是返回的Project，一个是RawPackage中的Project；
	// 返回*Project才能保证这是一个Project对象；
	// project.initRawPackage()
	// project.rawPkg = project.getPackage("", true)
	return GlobalProject
}
