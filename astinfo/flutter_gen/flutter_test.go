package flutter_gen

import (
	"strings"
	"testing"

	"github.com/wanjm/gos/astinfo"
)

func TestGenDTOOutput(t *testing.T) {
	gen := NewFlutterGen()
	statusField := newTestField("Status", "status", astinfo.GetRawType("int"))
	statusField.Comment.EnumName = "Status"
	s := &astinfo.Struct{
		StructName: "SampleDTO",
		Fields: []*astinfo.Field{
			newTestField("Name", "name", astinfo.GetRawType("string")),
			newTestField("ID", "_id", astinfo.GetRawType("string")),
			newTestField("CreatedAt", "createdAt", astinfo.GetRawType("int64")),
			statusField,
		},
	}
	mp := &astinfo.MainProject{
		EnumMap: map[string]*astinfo.EnumDef{
			"Status": {
				Name:      "Status",
				ValueKind: "int",
				Members: []astinfo.EnumMember{
					{MemberName: "Unknown", Value: "0"},
				},
			},
		},
	}

	got := strings.TrimSpace(gen.genDTO(s, mp))
	want := strings.TrimSpace(`// ["name","id","createdAt","status"]
class SampleDTO extends JSONParameter {
 /// 
 String name;
 /// 
 String id;
 /// 
 DateTime createdAt;
 /// 
 Status status;


  SampleDTO({
  this.name = "",
  this.id = "",
  required this.createdAt ,
  this.status = Status.unknown,

  });


  factory SampleDTO.fromJson(Map<String, dynamic> json) {
    return SampleDTO(
   name: json['name'] as String? ?? "",
   id: json['_id'] as String? ?? "",
   createdAt: DateTime.fromMillisecondsSinceEpoch((json['createdAt'] as num? ?? 0).toInt()),
   status: Status.fromValue((json['status'] as num?)?.toInt()) ?? Status.unknown,
    );
  }

  @override
  Map<String, dynamic> toJson() {
    return {
      "name": name,
      "_id": id,
      "createdAt": createdAt,
      "status": status.value,
    };
  }
}`)
	if got != want {
		t.Fatalf("genDTO output mismatch\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestNewDTOFieldNormalCase(t *testing.T) {
	gen := NewFlutterGen()
	field := newTestField("Name", "name", astinfo.GetRawType("string"))

	dto, ok := gen.newDTOField(field, nil)
	if !ok {
		t.Fatal("expected field to be generated")
	}
	assertDTOField(t, dto, DTOField{
		Name:         "name",
		JsonKey:      "name",
		DartType:     "String",
		DefaultValue: "\"\"",
		ParseString:  "json['name'] as String? ?? \"\"",
		ToJsonExpr:   "name",
	})
}

func TestNewDTOFieldMongoIDCase(t *testing.T) {
	gen := NewFlutterGen()
	field := newTestField("ID", "_id", astinfo.GetRawType("string"))

	dto, ok := gen.newDTOField(field, nil)
	if !ok {
		t.Fatal("expected field to be generated")
	}
	assertDTOField(t, dto, DTOField{
		Name:         "id",
		JsonKey:      "_id",
		DartType:     "String",
		DefaultValue: "\"\"",
		ParseString:  "json['_id'] as String? ?? \"\"",
		ToJsonExpr:   "id",
	})
}

func TestNewDTOFieldEpochMillisTimeCase(t *testing.T) {
	gen := NewFlutterGen()
	field := newTestField("CreatedAt", "createdAt", astinfo.GetRawType("int64"))

	dto, ok := gen.newDTOField(field, nil)
	if !ok {
		t.Fatal("expected field to be generated")
	}
	assertDTOField(t, dto, DTOField{
		Name:         "createdAt",
		JsonKey:      "createdAt",
		DartType:     "DateTime",
		DefaultValue: "DateTime.fromMillisecondsSinceEpoch(0)",
		ParseString:  "DateTime.fromMillisecondsSinceEpoch((json['createdAt'] as num? ?? 0).toInt())",
		ToJsonExpr:   "createdAt",
		Required:     true,
	})
}

func TestNewDTOFieldEnumCase(t *testing.T) {
	gen := NewFlutterGen()
	field := newTestField("Status", "status", astinfo.GetRawType("int"))
	field.Comment.EnumName = "Status"
	mp := &astinfo.MainProject{
		EnumMap: map[string]*astinfo.EnumDef{
			"Status": {
				Name:      "Status",
				ValueKind: "int",
				Members: []astinfo.EnumMember{
					{MemberName: "Unknown", Value: "0"},
				},
			},
		},
	}

	dto, ok := gen.newDTOField(field, mp)
	if !ok {
		t.Fatal("expected field to be generated")
	}
	assertDTOField(t, dto, DTOField{
		Name:         "status",
		JsonKey:      "status",
		DartType:     "Status",
		DefaultValue: "Status.unknown",
		ParseString:  "Status.fromValue((json['status'] as num?)?.toInt()) ?? Status.unknown",
		ToJsonExpr:   "status.value",
	})
}

func TestNewDTOFieldByteArrayCase(t *testing.T) {
	gen := NewFlutterGen()
	field := newTestField("Payload", "payload", &astinfo.ArrayType{Typer: astinfo.GetRawType("byte")})

	dto, ok := gen.newDTOField(field, nil)
	if !ok {
		t.Fatal("expected field to be generated")
	}
	assertDTOField(t, dto, DTOField{
		Name:         "payload",
		JsonKey:      "payload",
		DartType:     "String",
		DefaultValue: "\"\"",
		ParseString:  "json['payload']?.toString() ?? \"\"",
		ToJsonExpr:   "payload",
	})
}

func newTestField(name, jsonKey string, typer astinfo.Typer) *astinfo.Field {
	return &astinfo.Field{
		FieldBasic: astinfo.FieldBasic{
			Name: name,
			Type: typer,
		},
		Tags: map[string]string{"json": jsonKey},
	}
}

func assertDTOField(t *testing.T, got, want DTOField) {
	t.Helper()
	if got.Name != want.Name {
		t.Fatalf("Name = %q, want %q", got.Name, want.Name)
	}
	if got.JsonKey != want.JsonKey {
		t.Fatalf("JsonKey = %q, want %q", got.JsonKey, want.JsonKey)
	}
	if got.DartType != want.DartType {
		t.Fatalf("DartType = %q, want %q", got.DartType, want.DartType)
	}
	if got.DefaultValue != want.DefaultValue {
		t.Fatalf("DefaultValue = %q, want %q", got.DefaultValue, want.DefaultValue)
	}
	if got.ParseString != want.ParseString {
		t.Fatalf("ParseString = %q, want %q", got.ParseString, want.ParseString)
	}
	if got.ToJsonExpr != want.ToJsonExpr {
		t.Fatalf("ToJsonExpr = %q, want %q", got.ToJsonExpr, want.ToJsonExpr)
	}
	if got.Required != want.Required {
		t.Fatalf("Required = %v, want %v", got.Required, want.Required)
	}
}
