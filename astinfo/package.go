package astinfo

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

// Package 表示一个Go包的基本信息
type Package struct {
	Name    string             // 包名称
	Module  string             // 所属模块全路径
	Structs map[string]*Struct // 包内结构体集合（key为结构体名称）
}

func (pkg *Package) Parse(path string) error {
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
			// pkg.parseMod(f, filename)
			_ = f
			fmt.Printf("Parsing file: %s\n", filename)
			// gofile := createGoFile(f, pkg, filename)
			// gofile.parseFile()
		}
	}
	return nil
}

func (pkg *Package) processTypeDecl(decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		if _, ok := typeSpec.Type.(*ast.StructType); ok {
			pkg.Structs[typeSpec.Name.Name] = &Struct{
				Name: typeSpec.Name.Name,
			}
		}
	}
}
