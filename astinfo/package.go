package astinfo

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"strings"
)

// Package 表示一个Go包的基本信息
type Package struct {
	Simple     bool                  // 简单解析，及仅解析包名
	Name       string                // 包名称
	Path       string                // 包所在目录的绝对路径
	Module     string                // 所属模块全路径
	Structs    map[string]*Struct    // 包内结构体集合（key为结构体名称）
	Interfaces map[string]*Interface // key是Interface 的Name
	// 由于采用了两层扫描，所以不再需要Types map了。直接调用get方法获取；
	Types     map[string]Typer // key是Type 的Name
	fset      *token.FileSet   // 记录fset，到时可以找到文件
	GlobalVar map[string]*VarField
	FunctionManager
	finshedParse bool
}

//  Package中不包含goSource，因为
// 1. package 仅仅是因为需要被引用时生成；
// 2. package 有多个goSource文件；
// 3. package 将来可以建立goSource和ast.File之间的映射关系来达到寻找关系的目的；

// getName
func (pkg *Package) GetName() string {
	name := pkg.Name
	// Module Name有可能跟包名不一样，
	// 本模块需要解析的包都是有name的。但是第三方的包可能就没有，需要从Module Name中解析
	// 等后续有需求了再做；
	if name == "" {
		fmt.Printf("failed to get name of pkg %s, use base name in path which is error\n", pkg.Module)
		name = filepath.Base(pkg.Module)
	}
	return name
}

// 采用遇到遇到不认识的import就先深度parse的方法；
func (pkg *Package) Parse() error {
	if pkg.finshedParse {
		return nil
	}
	defer func() {
		pkg.finshedParse = true
	}()
	path := pkg.Path
	pkg.fset = token.NewFileSet()
	// 这里取绝对路径，方便打印出来的语法树可以转跳到编辑器
	packageMap, err := parser.ParseDir(pkg.fset, path, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		log.Printf("parse package %s failed %s", pkg.Module, err.Error())
		return nil
	}
	// 先简单扫描；仅扫描定义，struct，interface，var；
	var fileMap = make(map[*ast.File]*Gosourse)
	for packName, pack := range packageMap {
		_ = packName
		for filename, f := range pack.Files {
			if strings.HasSuffix(filename, "_test.go") {
				continue
			}
			gofile := NewGosourse(f, pkg, filename)
			gofile.ParseTop()
			fileMap[f] = gofile

		}
	}
	if pkg.Simple {
		return nil
	}
	fmt.Printf("Parsing package: %s\n", path)
	//再细化扫描；
	for packName, pack := range packageMap {
		_ = packName
		// fmt.Printf("begin parse %s with %s\n", packName, path)
		for _, f := range pack.Files {
			gofile := fileMap[f]
			if gofile != nil {
				gofile.Parse()
			}
		}
	}
	return nil
}

// NewPackage creates a new Package instance with the given module path
func NewPackage(module string, simple bool, absPath string) *Package {
	// Extract package name from module path
	return &Package{
		Module:     module,
		Simple:     simple,
		Path:       absPath,
		Structs:    make(map[string]*Struct),
		Interfaces: make(map[string]*Interface),
		GlobalVar:  make(map[string]*VarField),
		Types:      make(map[string]Typer),
	}
}

func NewSysPackage(module string) *Package {
	return &Package{
		Module:       module,
		Simple:       true,
		Path:         "",
		finshedParse: true,
		Name:         filepath.Base(module),
		Structs:      make(map[string]*Struct),
		Types:        make(map[string]Typer),
	}
}

func (pkg *Package) Getstruct(name string) *Struct {
	return pkg.Structs[name]
}

// findStruct
func (pkg *Package) FindStruct(name string) *Struct {
	class := pkg.Getstruct(name)
	if class == nil {
		class = NewStruct(name, pkg)
		pkg.Structs[name] = class
		pkg.Types[name] = class
	}
	return class
}
func (pkg *Package) GetTyper(name string) Typer {
	return pkg.Types[name]
}

// GetInterface
func (pkg *Package) GetInterface(name string) *Interface {
	return pkg.Interfaces[name]
}

// findInterface
func (pkg *Package) FindInterface(name string) *Interface {
	class := pkg.GetInterface(name)
	if class == nil {
		class = NewInterface(name, pkg)
		pkg.Interfaces[name] = class
		pkg.Types[name] = class
	}
	return class
}

func SimplePackage(module, name string) *Package {
	return &Package{
		Module: module,
		Name:   name,
	}
}
