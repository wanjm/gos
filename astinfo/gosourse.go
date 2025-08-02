package astinfo

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type Gosourse struct {
	Path    string
	Pkg     *Package
	File    *ast.File
	Imports map[string]string //key package name,value是包路径；
}

var knownowType = map[string]bool{}

// 解析全局变量和 struct，interface
func (g *Gosourse) getGenDeclParser(genDecl *ast.GenDecl) (parser Parser) {
	switch genDecl.Tok {
	case token.VAR:
		typeSpec := genDecl.Specs[0].(*ast.ValueSpec)
		parser = NewVarFieldHelper(typeSpec, g)
	case token.TYPE:
		typeSpec := genDecl.Specs[0].(*ast.TypeSpec)
		// 仅关注结构体，暂时不考虑接口
		switch typeType := typeSpec.Type.(type) {
		case *ast.InterfaceType:
			// fmt.Printf("InterfaceType %s %s\n", typeSpec.Name.Name, g.Path)
			class := g.Pkg.FindInterface(typeSpec.Name.Name)
			class.GoSource = g
			class.initGenDecl(genDecl)
			parser = class
		case *ast.StructType:
			// fmt.Printf("StructType %s %s\n", typeSpec.Name.Name, g.Path)
			class := g.Pkg.FindStruct(typeSpec.Name.Name)
			class.goSource = g
			class.initGenDecl(genDecl)
			parser = class
		case *ast.SelectorExpr:
			pkgName := typeSpec.Type.(*ast.SelectorExpr).X.(*ast.Ident).Name
			class := typeSpec.Type.(*ast.SelectorExpr).Sel.Name
			pkg := GlobalProject.FindPackage(g.Imports[pkgName])
			//TODO
			// fmt.Printf("define alias type %s => %s %s  \n", typeSpec.Name.Name, pkgName, class)
			typer := pkg.GetTyper(class)
			g.Pkg.Types[typeSpec.Name.Name] = typer
		default:
			var typeName = fmt.Sprintf("%T", typeType)
			if _, ok := knownowType[typeName]; !ok {
				knownowType[typeName] = true
				fmt.Printf("unknown type '%s' in '%T'\n", typeSpec.Name.Name, typeType)
			}
		}
	}
	return
}

// 解析函数和方法
func (g *Gosourse) getFuncDeclParser(funcDecl *ast.FuncDecl) Parser {
	switch funcDecl.Recv {
	case nil:
		return NewFunctionParserHelper(funcDecl, g)
	default:
		return NewMethod(funcDecl, g)
	}
}
func (g *Gosourse) ParseTop() bool {
	for _, c := range g.File.Comments {
		for _, line := range c.List {
			if strings.Contains(line.Text, "//go:build ignore") {
				return false
			}
		}
	}
	// TODO: 需要添加日志级别，再打印日志
	// fmt.Printf("Parsing file: %s name: %s %s\n", g.Path, g.Pkg.Name, g.Pkg.Module)
	g.parseImport(g.File.Imports)
	decls := g.File.Decls
	for i := 0; i < len(decls); i++ {
		switch decl := decls[i].(type) {
		case *ast.GenDecl:
			g.getGenDeclParser(decl)
		}
	}
	return true
}
func (g *Gosourse) Parse() error {
	decls := g.File.Decls
	for i := 0; i < len(decls); i++ {
		var parser Parser
		switch decl := decls[i].(type) {
		case *ast.GenDecl:
			parser = g.getGenDeclParser(decl)
		case *ast.FuncDecl:
			parser = g.getFuncDeclParser(decl)
		}
		if parser != nil {
			parser.Parse()
		}
	}
	// 方法体为空
	return nil
}

func NewGosourse(file *ast.File, pkg *Package, path string) *Gosourse {
	return &Gosourse{
		File:    file,
		Pkg:     pkg,
		Path:    path,
		Imports: map[string]string{},
	}
}

// 解析go文件的Import字段，如果有modeName直接使用，否则用pathValue的文件名；
// 注意此处可能有错误，因为有些package的模块名不是路径的最后一位；
// 此时只能通过解析原package文件才能解决；否则后面getImportPath就找不到了
func (goFile *Gosourse) parseImport(imports []*ast.ImportSpec) {
	for _, importSpec := range imports {
		var name string
		modulePath := importSpec.Path.Value
		pathValue := strings.Trim(modulePath, string(modulePath[0]))
		if pathValue == "C" {
			continue
		}
		if importSpec.Name != nil {
			name = importSpec.Name.Name
		} else {
			// 如果没有带名字，则从Package中寻找，此处是否可能该Package还没有被解析呢？
			pkg := GlobalProject.FindPackage(pathValue)
			name = pkg.GetName()
		}
		// pkg := goFile.pkg.Project.getPackage(pathValue, true)
		// 此处是第三方package，也可能是本项目的尚未被解析的工程，其modeName为空，先补一个；
		// 主要是为了解决package的ModeName不是其path的最后的baseName的情况
		// if len(pkg.modName) == 0 {
		// 	pkg.modName = name
		// }
		goFile.Imports[name] = pathValue
	}
}
