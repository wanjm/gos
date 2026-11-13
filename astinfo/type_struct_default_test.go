package astinfo

import (
	"strings"
	"testing"

	"github.com/wanjm/gos/astbasic"
)

// TestGenConstructCodeEmbeddedDefault 覆盖 request 构造时的匿名嵌入默认值。
// 嵌入结构体里带 default 的原始字段会写入字面量，没有 default 的嵌入结构体保持省略。
func TestGenConstructCodeEmbeddedDefault(t *testing.T) {
	pkg := &Package{PkgBasic: astbasic.PkgBasic{Name: "schema", ModPath: "example.com/schema"}}
	createReq := &Struct{
		StructName: "CreateReq",
		GoSource:   &Gosourse{Pkg: pkg},
		Fields: []*Field{
			rawDefaultField("PriceUserNum", "int", "6"),
			rawDefaultField("OpenComment", "bool", "true"),
			rawDefaultField("Title", "string", ""),
		},
	}
	outer := &Struct{
		StructName: "CreateByGroupReq",
		GoSource:   &Gosourse{Pkg: pkg},
		Fields: []*Field{
			embeddedField(createReq),
			namedStructField(pkg, "Extra"),
		},
	}

	got := outer.GenConstructCode(astbasic.CreateGenedFile("servlet"), false)
	want := strings.Join([]string{
		"schema.CreateByGroupReq{",
		"CreateReq:schema.CreateReq{",
		"PriceUserNum:6,",
		"OpenComment:true,",
		"},",
		"}",
	}, "\n")
	if got != want {
		t.Fatalf("embedded default =\n%s\nwant\n%s", got, want)
	}

	plain := &Struct{
		StructName: "Plain",
		GoSource:   &Gosourse{Pkg: pkg},
		Fields:     []*Field{rawDefaultField("N", "int", "")},
	}
	noDefault := &Struct{
		StructName: "Outer",
		GoSource:   &Gosourse{Pkg: pkg},
		Fields:     []*Field{embeddedField(plain)},
	}
	got = noDefault.GenConstructCode(astbasic.CreateGenedFile("servlet"), false)
	want = "schema.Outer{\n}"
	if got != want {
		t.Fatalf("embed without default =\n%s\nwant\n%s", got, want)
	}
}

// rawDefaultField 造一个带 default 标签的原始字段。
// def 为空时不写 default，构造时该字段被省略。
func rawDefaultField(name, typeName, def string) *Field {
	field := NewSimpleField(GetRawType(typeName), name)
	field.Tags = map[string]string{}
	if def != "" {
		field.Tags[DEFAULT] = def
	}
	return field
}

// embeddedField 造一个匿名嵌入字段。
// 字段名为空，构造时按被嵌入结构体的类型名写入。
func embeddedField(st *Struct) *Field {
	return &Field{FieldBasic: FieldBasic{Type: st}}
}

// namedStructField 造一个具名的结构体字段。
// request 构造不展开这种字段，即使它内部带 default。
func namedStructField(pkg *Package, name string) *Field {
	inner := &Struct{
		StructName: name,
		GoSource:   &Gosourse{Pkg: pkg},
		Fields:     []*Field{rawDefaultField("N", "int", "1")},
	}
	field := NewSimpleField(inner, name)
	field.Tags = map[string]string{}
	return field
}
