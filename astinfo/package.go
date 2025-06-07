package astinfo

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
)

// Package 表示一个Go包的基本信息
type Package struct {
	name    string             // 包名称
	Path    string             // 包所在目录的绝对路径
	Module  string             // 所属模块全路径
	Structs map[string]*Struct // 包内结构体集合（key为结构体名称）
}

// getName
func (pkg *Package) GetName() string {
	name := pkg.name
	// Module Name有可能跟包名不一样，
	// 本模块需要解析的包都是有name的。但是第三方的包可能就没有，需要从Module Name中解析
	// 等后续有需求了再做；
	if name == "" {
		name = filepath.Base(pkg.Module)
	}
	return name
}
func (pkg *Package) Parse() error {
	path := pkg.Path
	fmt.Printf("Parsing package: %s\n", path)
	fset := token.NewFileSet()
	// 这里取绝对路径，方便打印出来的语法树可以转跳到编辑器
	packageMap, err := parser.ParseDir(fset, path, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		log.Printf("parse %s failed %s", path, err.Error())
		return nil
	}
	for packName, pack := range packageMap {
		_ = packName
		// fmt.Printf("begin parse %s with %s\n", packName, path)
		for filename, f := range pack.Files {
			gofile := NewGosourse(f, pkg, filename)
			gofile.Parse()
		}
	}
	return nil
}

// NewPackage creates a new Package instance with the given module path
func NewPackage(module string) *Package {
	// Extract package name from module path
	return &Package{
		Module:  module,
		Structs: make(map[string]*Struct),
	}
}

func (pkg *Package) Gettruct(name string) *Struct {
	return pkg.Structs[name]
}

// findStruct
func (pkg *Package) FindStruct(name string) *Struct {
	class := pkg.Gettruct(name)
	if class == nil {
		class = NewStruct(name, pkg)
		pkg.Structs[name] = class
	}
	return class
}
