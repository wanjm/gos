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
}

func (p *Project) ParseModule() error {
	modPath := filepath.Join(p.Path, "go.mod")
	data, err := os.ReadFile(modPath)
	if err != nil {
		return fmt.Errorf("error reading go.mod: %w", err)
	}

	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return fmt.Errorf("error parsing go.mod: %w", err)
	}

	p.Module = modFile.Module.Mod.Path
	return nil
}

func (p *Project) Parse() error {
	if err := p.ParseModule(); err != nil {
		return err
	}

	p.Packages = make(map[string]*Package)

	err := filepath.WalkDir(p.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if err := p.ParsePackage(path); err != nil {
				return fmt.Errorf("error parsing package at %s: %w", path, err)
			}
		}
		return nil
	})

	return err
}

func (p *Project) ParsePackage(dir string) error {
	relPath, err := filepath.Rel(p.Path, dir)
	if err != nil {
		return err
	}

	pkgPath := filepath.Join(p.Module, relPath)
	pkg := &Package{
		Name:    filepath.Base(relPath),
		Module:  p.Module,
		Structs: make(map[string]*Struct),
	}

	if err := pkg.Parse(dir); err != nil {
		return fmt.Errorf("package parse error: %w", err)
	}

	p.Packages[pkgPath] = pkg
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
    
    // Extract package name from module path
    _, name := filepath.Split(module)
    if name == "" {
        name = "main"
    }
    
    newPkg := &Package{
        Name:    name,
        Module:  module,
        Structs: make(map[string]*Struct),
    }
    
    p.Packages[module] = newPkg
    return newPkg
}
