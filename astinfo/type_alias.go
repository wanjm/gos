package astinfo

type Alias struct {
	Equal bool
	Name  string
	Pkg   *Package
	Typer
}

func NewAlias(name string, pkg *Package, equal bool) *Alias {
	alias := &Alias{
		Equal: equal,
		Name:  name,
		Pkg:   pkg,
	}
	pkg.Types[name] = alias
	return alias
}

// RefName 实现Typer接口的RefName方法
func (a *Alias) RefName(genFile *GenedFile) string {
	if genFile == nil || genFile.pkg == a.Pkg {
		return a.Name
	}
	impt := genFile.GetImport(a.Pkg)
	return impt.Name + "." + a.Name
}

// IDName 实现Typer接口的IDName方法
func (a *Alias) IDName() string {
	return a.Pkg.Module + "." + a.Name
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
	return nil
}
