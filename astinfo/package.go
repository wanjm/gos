package astinfo

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"
)

// Package 表示一个Go包的基本信息
type Package struct {
	Name    string             // 包名称
	Module  string             // 所属模块全路径
	Structs map[string]*Struct // 包内结构体集合（key为结构体名称）
}

func (pkg *Package) Parse(dir string) error {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes,
		Dir:  dir,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return fmt.Errorf("failed to load package: %w", err)
	}

	for _, file := range pkgs[0].Syntax {
		for _, decl := range file.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				pkg.processTypeDecl(genDecl)
			}
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
