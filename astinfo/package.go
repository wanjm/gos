package astinfo

import (
	"go/ast"
	"go/parser"
	"go/token"
)

// Package 表示一个Go包的结构信息
type Package struct {
	Name    string             // 包名称（与目录名一致）
	Module  string             // 包的全路径（如"github.com/user/project/pkg"）
	Structs map[string]*Struct // 包内定义的结构体集合
}

// NewPackage 创建并初始化一个新的Package实例
func NewPackage() *Package {
	return &Package{
		Structs: make(map[string]*Struct),
	}
}

// Parse 解析指定目录下的Go包
func (p *Package) Parse(dirPath string) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dirPath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, astPkg := range pkgs {
		p.Name = astPkg.Name

		for _, file := range astPkg.Files {
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
					for _, spec := range genDecl.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							if _, ok := typeSpec.Type.(*ast.StructType); ok {
								p.Structs[typeSpec.Name.Name] = NewStruct()
							}
						}
					}
				}
			}
		}
	}
	return nil
}
