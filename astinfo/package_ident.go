package astinfo

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
)

// LookupPackageIdent finds a parsed package named pkgName that declares ident
// as an exported const. Used for @gos permission=pkg.Const so servlet.gen.go
// can import that package without the servlet source file importing it
// (the name lives in a comment).
// 从所有的package中找出packageName为pkgName的，然后将这些pkg按照modPath排序；
// 然后遍历这些pkg，如果pkg中声明了该导出 const，则收集这个pkg；
func (mp *MainProject) LookupPackageIdent(pkgName, ident string) (*Package, error) {
	if !token.IsExported(ident) {
		return nil, fmt.Errorf("%s.%s is not an exported const", pkgName, ident)
	}
	if mp == nil || mp.Packages == nil {
		return nil, fmt.Errorf("package %s not found for %s.%s", pkgName, pkgName, ident)
	}
	var named []*Package
	for _, pkg := range mp.Packages {
		if pkg == nil {
			continue
		}
		name := pkg.Name
		if name == "" {
			name = filepath.Base(pkg.ModPath)
		}
		if name != pkgName {
			continue
		}
		named = append(named, pkg)
	}
	sort.Slice(named, func(i, j int) bool {
		return named[i].ModPath < named[j].ModPath
	})
	if len(named) == 0 {
		return nil, fmt.Errorf("package %s not found for %s.%s", pkgName, pkgName, ident)
	}
	var declared []*Package
	for _, pkg := range named {
		if packageDeclaresConst(pkg, ident) {
			declared = append(declared, pkg)
		}
	}
	if len(declared) == 1 {
		return declared[0], nil
	}
	if len(declared) > 1 {
		return nil, fmt.Errorf("ambiguous %s.%s in packages %s", pkgName, ident, packageModPaths(declared))
	}
	if len(named) == 1 {
		return nil, fmt.Errorf("%s.%s is not an exported const in package %s", pkgName, ident, named[0].ModPath)
	}
	return nil, fmt.Errorf("%s.%s is not an exported const in packages named %s: %s", pkgName, ident, pkgName, packageModPaths(named))
}

func packageModPaths(pkgs []*Package) string {
	paths := make([]string, len(pkgs))
	for i, pkg := range pkgs {
		paths[i] = pkg.ModPath
	}
	return strings.Join(paths, ", ")
}

func packageDeclaresConst(pkg *Package, ident string) bool {
	for _, enum := range pkg.Enums {
		if enum == nil {
			continue
		}
		for _, m := range enum.Members {
			if m.GoName == ident {
				return true
			}
		}
	}
	for _, f := range pkg.Files {
		if astFileDeclaresConst(f, ident) {
			return true
		}
	}
	return false
}

func astFileDeclaresConst(f *ast.File, ident string) bool {
	if f == nil {
		return false
	}
	for _, decl := range f.Decls {
		d, ok := decl.(*ast.GenDecl)
		if !ok || d.Tok != token.CONST {
			continue
		}
		for _, spec := range d.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, n := range vs.Names {
				if n.Name == ident {
					return true
				}
			}
		}
	}
	return false
}
