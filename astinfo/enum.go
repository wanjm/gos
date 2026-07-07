package astinfo

import (
	"fmt"
	"strings"

	"github.com/wanjm/gos/astbasic"
)

const (
	EnumKey        = "enum"
	DisplayWordKey = "displayWord"
)

// EnumMember is one const value in a @gos enum block.
type EnumMember struct {
	GoName      string
	MemberName  string
	Value       string // codegen literal snippet, e.g. "2", "\"hello\"", "true"
	Comment     string // inline line comment on the const member
	DisplayWord string
}

// EnumDef describes a const block annotated with @gos enum.
type EnumDef struct {
	Name      string
	GoSource  *Gosourse
	ValueKind string // "int" | "string" | "bool"
	Members   []EnumMember
}

func enumMemberName(goName string) string {
	if idx := strings.Index(goName, "_"); idx >= 0 {
		return goName[idx+1:]
	}
	return goName
}

func registerEnum(enum *EnumDef) {
	pkg := enum.GoSource.Pkg
	if pkg.Enums == nil {
		pkg.Enums = make(map[string]*EnumDef)
	}
	pkg.Enums[enum.Name] = enum
	if GlobalProject == nil {
		return
	}
	if GlobalProject.EnumMap == nil {
		GlobalProject.EnumMap = make(map[string]*EnumDef)
	}
	if _, exists := GlobalProject.EnumMap[enum.Name]; exists {
		fmt.Printf("WARNING: duplicate enum name %q (keep first)\n", enum.Name)
		return
	}
	GlobalProject.EnumMap[enum.Name] = enum
}

func (mp *MainProject) LookupEnum(name string) *EnumDef {
	if mp == nil || name == "" {
		return nil
	}
	return mp.EnumMap[name]
}

func (e *EnumDef) DartValueType() string {
	switch e.ValueKind {
	case "string":
		return "String"
	case "bool":
		return "bool"
	default:
		return "int"
	}
}

func (e *EnumDef) DartDefaultValue() string {
	if len(e.Members) == 0 {
		return "null"
	}
	return e.Name + "." + e.Members[0].DartMemberName()
}

func (e *EnumDef) DartParseExpr(jsonKey string) string {
	expr := dartJSONExprForKind(e.ValueKind, jsonKey)
	return e.Name + ".fromValue(" + expr + ")"
}

func (e *EnumDef) DartToJSONExpr(fieldName string) string {
	return fieldName + ".value"
}

func dartJSONExprForKind(kind, jsonKey string) string {
	switch kind {
	case "string":
		return "json['" + jsonKey + "'] as String?"
	case "bool":
		return "json['" + jsonKey + "'] as bool?"
	default:
		return "(json['" + jsonKey + "'] as num?)?.toInt()"
	}
}

func (e *EnumDef) TSValueType() string {
	switch e.ValueKind {
	case "string":
		return "string"
	case "bool":
		return "boolean"
	default:
		return "number"
	}
}

func (m *EnumMember) DartMemberName() string {
	return astbasic.FirstLower(m.MemberName)
}

func (m *EnumMember) TSMemberName() string {
	return m.MemberName
}

// DartValueLiteral returns the value expression pasted into generated Dart code.
func (m *EnumMember) DartValueLiteral(kind string) string {
	switch kind {
	case "string":
		return goStringSnippetToDart(m.Value)
	case "bool", "int":
		return m.Value
	default:
		return m.Value
	}
}

// TSValueLiteral returns the value expression pasted into generated TypeScript code.
func (m *EnumMember) TSValueLiteral(kind string) string {
	switch kind {
	case "string":
		return goStringSnippetToTS(m.Value)
	case "bool", "int":
		return m.Value
	default:
		return m.Value
	}
}

func goStringSnippetToDart(snippet string) string {
	unquoted := strings.Trim(snippet, `"`)
	return "'" + strings.ReplaceAll(unquoted, "'", "\\'") + "'"
}

func goStringSnippetToTS(snippet string) string {
	unquoted := strings.Trim(snippet, `"`)
	return "'" + strings.ReplaceAll(unquoted, "'", "\\'") + "'"
}

func escapeDartString(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}

func escapeTSString(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}
