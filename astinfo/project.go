package astinfo

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

	// 直接读取第一行解析module值
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return fmt.Errorf("go.mod is empty")
	}

	firstLine := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(firstLine, "module ") {
		return fmt.Errorf("invalid go.mod format, missing module declaration")
	}

	p.Module = strings.TrimSpace(firstLine[7:])
	return nil
}

func (p *Project) Parse() error {
	if err := p.ParseModule(); err != nil {
		return err
	}

	p.Packages = make(map[string]*Package)

	err := filepath.WalkDir(p.Path, func(path string, d fs.DirEntry, err error) error {
		//path是全路径
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

// dir是pacakge 所在的全路径
func (p *Project) ParsePackage(dir string) error {
	relPath, err := filepath.Rel(p.Path, dir)
	if err != nil {
		return err
	}

	pkgPath := filepath.Join(p.Module, relPath)
	pkg := p.GetPackage(pkgPath)
	if err := pkg.Parse(dir); err != nil {
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
