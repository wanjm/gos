package astinfo

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path"
	"path/filepath"
	"strings"
)

// Package 表示一个Go包的基本信息
type Package struct {
	Simple  bool               // 简单解析，及仅解析包名
	Name    string             // 包名称
	Path    string             // 包所在目录的绝对路径
	Module  string             // 所属模块全路径
	Structs map[string]*Struct // 包内结构体集合（key为结构体名称）
	// Interfaces map[string]*Interface // key是Interface 的Name
	// 由于采用了两层扫描，所以不再需要Types map了。直接调用get方法获取；
	Types     map[string]Typer // key是Type 的Name
	fset      *token.FileSet   // 记录fset，到时可以找到文件
	GlobalVar map[string]*VarField
	WaitTyper map[string][]*Typer // 有些类型先被使用，再定义，此时在此处将内容缓存下载，最后统一解析；
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
	var a complex128
	_ = a
	path := pkg.Path
	defer func() {
		pkg.finshedParse = true
		for name, alias := range pkg.WaitTyper {
			typer := pkg.GetTyper(name)
			if typer != nil {
				for _, typer1 := range alias {
					*typer1 = typer
				}
			} else {
				fmt.Printf("Error: failed to get %s.%s when parse finish\n", pkg.Path, name)
			}
		}
		// fmt.Printf("finished Parsing package: %s\n", path)
	}()
	pkg.fset = token.NewFileSet()
	// 这里取绝对路径，方便打印出来的语法树可以转跳到编辑器
	// fmt.Printf("Parsing package: %s\n", path)
	packageMap, err := parser.ParseDir(pkg.fset, path, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		log.Printf("parse package %s failed %s", pkg.Module, err.Error())
		return nil
	}
	// 先简单扫描；仅扫描定义，struct，interface，var；
	var fileMap = make(map[*ast.File]*Gosourse)
	// 一个目录下可能有多
	// var count = 0
	// for packName, _ := range packageMap {
	// 	if strings.HasSuffix(packName, "_test") {
	// 		continue
	// 	}
	// 	count++
	// }
	// if count > 1 {
	// 	fmt.Printf("more than one package\n")
	// }
	for packName, pack := range packageMap {
		_ = packName
		// 先跳过test,后续再说；
		if strings.HasSuffix(packName, "_test") {
			continue
		}
		for filename, f := range pack.Files {
			if strings.HasSuffix(filename, "_test.go") {
				continue
			}
			gofile := NewGosourse(f, pkg, filename)
			if gofile.ParseTop() {
				if pkg.Name == "" {
					pkg.Name = packName
				} else {
					// 本代码默认一个目录下仅有一个package；
					// 1. 前面的代码已经跳过了test；
					// 2. 如果还有其他的packge，则其不应该参与解析；ParseTop应该会跳过；
					if pkg.Name != packName {
						fmt.Printf("package name not equal %s %s\n", pkg.Name, packName)
					}
				}
				fileMap[f] = gofile
			}
		}
	}
	if pkg.Simple {
		return nil
	}
	// TODO
	//再细化扫描；
	for _, gofile := range fileMap {
		gofile.Parse()
	}

	return nil
}

// NewPackage creates a new Package instance with the given module path
func NewPackage(module string, simple bool, absPath string) *Package {
	// Extract package name from module path
	return &Package{
		Module:  module,
		Simple:  simple,
		Path:    absPath,
		Structs: make(map[string]*Struct),
		// Interfaces: make(map[string]*Interface),
		GlobalVar: make(map[string]*VarField),
		Types:     make(map[string]Typer),
		WaitTyper: make(map[string][]*Typer),
		// finshedParse: simple,
	}
}

func NewSysPackage(module string) *Package {
	return NewPackage(module, true, path.Join("/opt/google/go/src", module))
}

func (pkg *Package) Getstruct(name string) *Struct {
	return pkg.Structs[name]
}

// findStruct
func (pkg *Package) FindStruct(name string) *Struct {
	class := pkg.Getstruct(name)
	if class == nil {
		class = NewStruct(name, pkg)
	}
	return class
}

func (pkg *Package) FillType(typeName string, typer *Typer) {
	res := pkg.GetTyper(typeName)
	if res == nil {
		pkg.WaitTyper[typeName] = append(pkg.WaitTyper[typeName], typer)
	} else {
		*typer = res
	}
}
func (pkg *Package) GetTyper(name string) Typer {
	return pkg.Types[name]
}

func SimplePackage(module, name string) *Package {
	return &Package{
		Module: module,
		Name:   name,
	}
}
