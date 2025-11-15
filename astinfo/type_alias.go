package astinfo

import "go/ast"

type Alias struct {
	Equal    bool
	Name     string // type Name = StructName
	astRoot  *ast.TypeSpec
	Gosourse *Gosourse
	Typer
}

func NewAlias(astRoot *ast.TypeSpec, g *Gosourse, equal bool) *Alias {
	alias := &Alias{
		Equal:    equal,
		Name:     astRoot.Name.Name,
		astRoot:  astRoot,
		Gosourse: g,
	}
	g.Pkg.Types[astRoot.Name.Name] = alias
	return alias
}

// RefName 实现Typer接口的RefName方法
func (a *Alias) RefName(genFile *GenedFile) string {
	if genFile == nil || a.Gosourse.Pkg.IsSame(genFile.Pkg) {
		return a.Name
	}
	impt := genFile.GetImport(&a.Gosourse.Pkg.PkgBasic)
	return impt.Name + "." + a.Name
}

// IDName 实现Typer接口的IDName方法
func (a *Alias) IDName() string {
	return a.Gosourse.Pkg.ModPath + "." + a.Name
}

// GenConstructCode 实现Constructor接口的GenConstructCode方法
func (a *Alias) GenConstructCode(genFile *GenedFile, wire bool) string {
	if a.Typer != nil {
		return a.Typer.GenConstructCode(genFile, wire)
	}
	// 如果没有基础类型，返回零值
	return "nil"
}

// Parse() error
func (a *Alias) Parse() error {
	var typeMap map[string]*Field
	if a.astRoot.TypeParams != nil {
		typeParameter := parseFields(a.astRoot.TypeParams.List, a.Gosourse, nil)
		typeMap = FieldListToMap(typeParameter)
	}
	a.Typer = parseType(a.astRoot.Type, a.Gosourse, typeMap)
	return nil
}
