package astinfo

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestLookupPackageIdentFromFileAST(t *testing.T) {
	src := `package user_permission

const (
	PermissionGroupUser       = 0
	PermissionTypeCreateAdmin = 1
)
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "const.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	pkg := NewPackage("example.com/mod/user_permission", false, "/tmp")
	pkg.Name = "user_permission"
	pkg.Files = map[string]*ast.File{"const.go": f}

	mp := &MainProject{
		Packages: map[string]*Package{pkg.ModPath: pkg},
	}

	got, err := mp.LookupPackageIdent("user_permission", "PermissionTypeCreateAdmin")
	if err != nil {
		t.Fatal(err)
	}
	if got != pkg {
		t.Fatalf("got %v want the user_permission package", got)
	}

	got, err = mp.LookupPackageIdent("user_permission", "PermissionGroupUser")
	if err != nil {
		t.Fatal(err)
	}
	if got != pkg {
		t.Fatalf("got %v want the user_permission package", got)
	}

	if _, err := mp.LookupPackageIdent("user_permission", "Missing"); err == nil {
		t.Fatal("expected error for missing ident")
	}
	if _, err := mp.LookupPackageIdent("other", "PermissionTypeCreateAdmin"); err == nil {
		t.Fatal("expected error for missing package")
	}
}

func TestLookupPackageIdentOnlyExportedConst(t *testing.T) {
	src := `package p

type FlagType int
var FlagVar = 1
func FlagFunc() {}
const unexportedFlag = 2
const FlagConst = 3
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "file.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	pkg := NewPackage("example.com/p", false, "/tmp")
	pkg.Name = "p"
	pkg.Files = map[string]*ast.File{"file.go": f}
	mp := &MainProject{
		Packages: map[string]*Package{pkg.ModPath: pkg},
	}

	got, err := mp.LookupPackageIdent("p", "FlagConst")
	if err != nil {
		t.Fatal(err)
	}
	if got != pkg {
		t.Fatalf("got %v want package p", got)
	}
	for _, ident := range []string{"FlagType", "FlagVar", "FlagFunc", "unexportedFlag"} {
		if _, err := mp.LookupPackageIdent("p", ident); err == nil {
			t.Fatalf("expected error for non-exported-const %s", ident)
		}
	}
}

func TestLookupPackageIdentAmbiguous(t *testing.T) {
	src := `package dup

const Flag = 1
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "const.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	a := NewPackage("example.com/a/dup", false, "/tmp/a")
	a.Name = "dup"
	a.Files = map[string]*ast.File{"const.go": f}
	b := NewPackage("example.com/b/dup", false, "/tmp/b")
	b.Name = "dup"
	b.Files = map[string]*ast.File{"const.go": f}
	mp := &MainProject{
		Packages: map[string]*Package{a.ModPath: a, b.ModPath: b},
	}
	if _, err := mp.LookupPackageIdent("dup", "Flag"); err == nil {
		t.Fatal("expected ambiguous error")
	}
}

func TestLookupPackageIdentFromEnum(t *testing.T) {
	pkg := NewPackage("example.com/mod/user_permission", false, "/tmp")
	pkg.Name = "user_permission"
	pkg.Enums = map[string]*EnumDef{
		"PermissionType": {
			Members: []EnumMember{{GoName: "PermissionTypeCreateAdmin"}},
		},
	}
	mp := &MainProject{
		Packages: map[string]*Package{pkg.ModPath: pkg},
	}
	got, err := mp.LookupPackageIdent("user_permission", "PermissionTypeCreateAdmin")
	if err != nil {
		t.Fatal(err)
	}
	if got != pkg {
		t.Fatalf("got %v want the user_permission package", got)
	}
}
