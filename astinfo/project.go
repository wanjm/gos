package astinfo

import (
	"fmt"
	"go/build"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// Project 表示一个Go项目的基本信息
type Project struct {
	Simple   bool
	Name     string // 项目名称
	ModPath  string // 项目模块名称（从go.mod解析）
	FilePath string // 项目根目录的绝对路径
	Require  []*modfile.Require
}

func (p *Project) ParseModule() error {
	modPath := filepath.Join(p.FilePath, "go.mod")
	data, err := os.ReadFile(modPath)
	if err != nil {
		return fmt.Errorf("error reading go.mod: %w", err)
	}

	modfile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return fmt.Errorf("error parsing go.mod: %w", err)
	}
	p.ModPath = modfile.Module.Mod.Path
	p.Require = modfile.Require
	// TODO
	// fmt.Printf("Module: %s\n", p.Module)
	return nil
}

func (p *Project) Parse() error {
	if err := p.ParseModule(); err != nil {
		return err
	}
	return p.ParseCode()
}
func (p *Project) ParseCode1() error {
	pkg, err := build.ImportDir(p.FilePath, 0)
	if err != nil {
		fmt.Printf("无法解析目录: %v\n", err)
		return nil
	}
	fmt.Print(pkg.Name, "\n")
	return nil
}
func (p *Project) ParseCode() error {
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
	err := filepath.WalkDir(p.FilePath, func(path string, d fs.DirEntry, err error) error {
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
	relPath, err := filepath.Rel(p.FilePath, dir)
	if err != nil {
		return err
	}
	relPath = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")
	// 用mode+相对路径，得到包全路径
	pkgPath := path.Join(p.ModPath, relPath)
	pkg := GlobalProject.FindPackage(pkgPath)
	pkg.Parse()
	// pkg.Simple = p.Simple
	// pkg.Path = dir
	// if err := pkg.Parse(); err != nil {
	// 	return fmt.Errorf("package parse error: %w", err)
	// }
	return nil
}
