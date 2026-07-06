package test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/wanjm/gos/astinfo"
	"github.com/wanjm/gos/astinfo/flutter_gen"
	"github.com/wanjm/gos/astinfo/ts_gen"
)

func parseConstEnums(t *testing.T, src string) map[string]*astinfo.EnumDef {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	pkg := astinfo.NewPackage("example.com/test", false, ".")
	pkg.Name = "testpkg"
	goSource := astinfo.NewGosourse(file, pkg, "fixture.go")
	astinfo.GlobalProject = &astinfo.MainProject{
		EnumMap: make(map[string]*astinfo.EnumDef),
	}
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		p := astinfo.NewConstBlockParser(genDecl, goSource)
		if err := p.Parse(); err != nil {
			t.Fatalf("parse const: %v", err)
		}
	}
	return astinfo.GlobalProject.EnumMap
}

func TestParseIntEnum(t *testing.T) {
	src := `package testpkg
// @gos enum=PackageType;
const (
	// @gos displayWord="普通课程";
	PackageType_Normal = 1 //普通课程
	PackageType_Combo  = 2 //组合课程
)`
	enums := parseConstEnums(t, src)
	enum := enums["PackageType"]
	if enum == nil {
		t.Fatal("PackageType enum not found")
	}
	if enum.ValueKind != "int" {
		t.Fatalf("ValueKind = %q, want int", enum.ValueKind)
	}
	if len(enum.Members) != 2 {
		t.Fatalf("members = %d, want 2", len(enum.Members))
	}
	if enum.Members[0].Value != "1" {
		t.Fatalf("first value = %q, want 1", enum.Members[0].Value)
	}
	if enum.Members[0].DisplayWord != "普通课程" {
		t.Fatalf("displayWord = %q", enum.Members[0].DisplayWord)
	}
	if enum.Members[1].DisplayWord != "组合课程" {
		t.Fatalf("displayWord fallback = %q", enum.Members[1].DisplayWord)
	}
}

func TestParseIotaEnum(t *testing.T) {
	src := `package testpkg
// @gos enum=Status;
const (
	Status_A = iota + 1 // 启用
	Status_B            // 停用
)`
	enums := parseConstEnums(t, src)
	enum := enums["Status"]
	if enum == nil {
		t.Fatal("Status enum not found")
	}
	if enum.Members[0].Value != "1" || enum.Members[1].Value != "2" {
		t.Fatalf("iota values = %q, %q", enum.Members[0].Value, enum.Members[1].Value)
	}
}

func TestParseStringEnum(t *testing.T) {
	src := `package testpkg
// @gos enum=AiAgentType;
const (
	AI_HWL_NORMAL = "hwlnormal" // 好未来通用
)`
	enums := parseConstEnums(t, src)
	enum := enums["AiAgentType"]
	if enum == nil {
		t.Fatal("AiAgentType enum not found")
	}
	if enum.ValueKind != "string" {
		t.Fatalf("ValueKind = %q, want string", enum.ValueKind)
	}
	if enum.Members[0].Value != `"hwlnormal"` {
		t.Fatalf("value snippet = %q, want \"hwlnormal\"", enum.Members[0].Value)
	}
}

func TestParseBoolEnum(t *testing.T) {
	src := `package testpkg
// @gos enum=Flag;
const (
	Flag_Yes = true  // 是
	Flag_No  = false // 否
)`
	enums := parseConstEnums(t, src)
	enum := enums["Flag"]
	if enum == nil {
		t.Fatal("Flag enum not found")
	}
	if enum.ValueKind != "bool" {
		t.Fatalf("ValueKind = %q, want bool", enum.ValueKind)
	}
	if enum.Members[0].Value != "true" || enum.Members[1].Value != "false" {
		t.Fatalf("bool values = %q, %q", enum.Members[0].Value, enum.Members[1].Value)
	}
}

func TestInferEnumNameFromPrefix(t *testing.T) {
	src := `package testpkg
// @gos enum;
const (
	PackageType_Normal = 1
	PackageType_Combo  = 2
)`
	enums := parseConstEnums(t, src)
	if enums["PackageType"] == nil {
		t.Fatal("expected inferred enum name PackageType")
	}
}

func TestFieldCommentEnumFromDoc(t *testing.T) {
	src := `package testpkg
type CourseInfo struct {
	// @gos enum=PackageType
	PackageType int8 ` + "`json:\"packageType\"`" + ` // 课程类型
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	pkg := astinfo.NewPackage("example.com/test", false, ".")
	pkg.Name = "testpkg"
	goSource := astinfo.NewGosourse(file, pkg, "fixture.go")
	typeSpec := file.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
	st := typeSpec.Type.(*ast.StructType)
	field := astinfo.NewField(st.Fields.List[0], goSource)
	field.Parse(nil)

	if field.Comment.EnumName != "PackageType" {
		t.Fatalf("EnumName = %q, want PackageType", field.Comment.EnumName)
	}
	if field.Comment.CommentText != "课程类型" {
		t.Fatalf("CommentText = %q, want 课程类型", field.Comment.CommentText)
	}
}

func TestFieldCommentEnumNotFromLineComment(t *testing.T) {
	src := `package testpkg
type CourseInfo struct {
	PackageType int8 ` + "`json:\"packageType\"`" + ` // @gos enum=PackageType
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	pkg := astinfo.NewPackage("example.com/test", false, ".")
	goSource := astinfo.NewGosourse(file, pkg, "fixture.go")
	typeSpec := file.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
	st := typeSpec.Type.(*ast.StructType)
	field := astinfo.NewField(st.Fields.List[0], goSource)
	field.Parse(nil)

	if field.Comment.EnumName != "" {
		t.Fatalf("EnumName should be empty when @gos is only on line comment, got %q", field.Comment.EnumName)
	}
}

func TestGenEnumDart(t *testing.T) {
	enum := &astinfo.EnumDef{
		Name:      "PackageType",
		ValueKind: "int",
		Members: []astinfo.EnumMember{
			{MemberName: "Normal", Value: "1", DisplayWord: "普通课程"},
			{MemberName: "Combo", Value: "2", DisplayWord: "组合课程"},
		},
	}
	out := flutter_gen.GenEnumDart(enum)
	if !strings.Contains(out, "normal(1, '普通课程')") {
		t.Fatalf("dart output missing member: %s", out)
	}
	if !strings.Contains(out, "String text() => displayWord") {
		t.Fatalf("dart output missing text(): %s", out)
	}
}

func TestGenEnumTS(t *testing.T) {
	enum := &astinfo.EnumDef{
		Name:      "PackageType",
		ValueKind: "int",
		Members: []astinfo.EnumMember{
			{MemberName: "Normal", Value: "1", DisplayWord: "普通课程"},
		},
	}
	out := ts_gen.GenEnumTS(enum)
	if !strings.Contains(out, "Normal = 1") {
		t.Fatalf("ts output missing member: %s", out)
	}
	if !strings.Contains(out, "function packageTypeText") {
		t.Fatalf("ts output missing text helper: %s", out)
	}

	strEnum := &astinfo.EnumDef{
		Name:      "AiAgentType",
		ValueKind: "string",
		Members: []astinfo.EnumMember{
			{MemberName: "HwlNormal", Value: `"hwlnormal"`, DisplayWord: "好未来"},
		},
	}
	strOut := ts_gen.GenEnumTS(strEnum)
	if !strings.Contains(strOut, "HwlNormal: 'hwlnormal'") {
		t.Fatalf("ts string enum output: %s", strOut)
	}
}
