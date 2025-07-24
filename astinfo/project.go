package astinfo

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// Project 表示一个Go项目的基本信息
type Project struct {
	Name   string // 项目名称
	Module string // 项目模块名称（从go.mod解析）
	Path   string // 项目根目录的绝对路径
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
	var goPath = os.Getenv("GOPATH")
	for _, req := range modfile.Require {
		if !req.Indirect {
			pkg := GlobalProject.FindPackage(req.Mod.Path)
			pkg.Path = path.Join(goPath, req.Mod.Path+req.Mod.Version)
		}
		fmt.Printf("Module: %s\n", req.Mod.Path)
	}
	fmt.Printf("Module: %s\n", p.Module)
	return nil
}

func (p *Project) Parse() error {
	if err := p.ParseModule(); err != nil {
		return err
	}
	// p.Packages = make(map[string]*Package)
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
	pkg := GlobalProject.FindPackage(pkgPath)
	pkg.Path = dir
	if err := pkg.Parse(); err != nil {
		return fmt.Errorf("package parse error: %w", err)
	}
	return nil
}
