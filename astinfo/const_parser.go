package astinfo

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

type constBlockComment struct {
	enumName string
	hasEnum  bool
}

func (c *constBlockComment) dealValuePair(key, value string) {
	value = strings.Trim(value, "\"")
	switch key {
	case EnumKey:
		c.hasEnum = true
		if value != "" {
			c.enumName = value
		}
	}
}

type constMemberComment struct {
	displayWord string
}

func (c *constMemberComment) dealValuePair(key, value string) {
	value = strings.Trim(value, "\"")
	switch key {
	case DisplayWordKey:
		c.displayWord = value
	}
}

type ConstBlockParser struct {
	genDecl  *ast.GenDecl
	goSource *Gosourse
}

func NewConstBlockParser(genDecl *ast.GenDecl, goSource *Gosourse) *ConstBlockParser {
	return &ConstBlockParser{
		genDecl:  genDecl,
		goSource: goSource,
	}
}

func (p *ConstBlockParser) Parse() error {
	blockComment := &constBlockComment{}
	parseComment(p.genDecl.Doc, blockComment)
	if !blockComment.hasEnum {
		return nil
	}

	enumName := blockComment.enumName
	if enumName == "" {
		enumName = inferEnumNameFromSpecs(p.genDecl.Specs)
	}
	if enumName == "" {
		fmt.Printf("WARNING: @gos enum block in %s has no name and no inferrable prefix\n", p.goSource.Path)
		return nil
	}

	var members []EnumMember
	var valueKind string
	var lastRhsExpr ast.Expr

	for i, spec := range p.genDecl.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		iotaVal := int64(i)

		kind := inferValueKind(vs)
		if kind == "" && lastRhsExpr != nil {
			kind = exprValueKind(lastRhsExpr)
		}
		if kind == "" && valueKind != "" {
			kind = valueKind
		}
		if kind == "" {
			fmt.Printf("WARNING: skip const in enum %q at %s: unsupported value kind\n", enumName, p.goSource.Path)
			continue
		}
		if valueKind == "" {
			valueKind = kind
		} else if valueKind != kind {
			fmt.Printf("WARNING: skip enum %q at %s: mixed value kinds %s and %s\n", enumName, p.goSource.Path, valueKind, kind)
			return nil
		}

		var values []string
		if len(vs.Values) == 0 {
			if lastRhsExpr != nil {
				snippet, ok := evalConstValue(lastRhsExpr, iotaVal, valueKind)
				if !ok {
					fmt.Printf("WARNING: skip enum %q at %s: cannot evaluate inherited const value\n", enumName, p.goSource.Path)
					continue
				}
				values = []string{snippet}
			} else if valueKind == "int" {
				values = []string{strconv.FormatInt(iotaVal, 10)}
			} else {
				fmt.Printf("WARNING: skip enum %q at %s: missing value for non-int const\n", enumName, p.goSource.Path)
				continue
			}
		} else {
			lastRhsExpr = vs.Values[len(vs.Values)-1]
			for _, valExpr := range vs.Values {
				snippet, ok := evalConstValue(valExpr, iotaVal, valueKind)
				if !ok {
					fmt.Printf("WARNING: skip enum %q at %s: cannot evaluate const value\n", enumName, p.goSource.Path)
					continue
				}
				values = append(values, snippet)
			}
		}

		for j, name := range vs.Names {
			if name.Name == "_" {
				continue
			}
			valueSnippet := values[0]
			if j < len(values) {
				valueSnippet = values[j]
			}
			memberComment := &constMemberComment{}
			parseComment(vs.Doc, memberComment)
			inlineComment := enumMemberInlineComment(vs.Comment)
			displayWord := memberComment.displayWord
			if displayWord == "" {
				displayWord = inlineComment
			}
			members = append(members, EnumMember{
				GoName:      name.Name,
				MemberName:  enumMemberName(name.Name),
				Value:       valueSnippet,
				Comment:     inlineComment,
				DisplayWord: displayWord,
			})
		}
	}

	if len(members) == 0 {
		return nil
	}

	registerEnum(&EnumDef{
		Name:      enumName,
		GoSource:  p.goSource,
		ValueKind: valueKind,
		Members:   members,
	})
	return nil
}

func inferEnumNameFromSpecs(specs []ast.Spec) string {
	for _, spec := range specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok || len(vs.Names) == 0 {
			continue
		}
		name := vs.Names[0].Name
		if name == "_" {
			continue
		}
		if idx := strings.Index(name, "_"); idx > 0 {
			return name[:idx]
		}
		return name
	}
	return ""
}

func inferValueKind(vs *ast.ValueSpec) string {
	if vs.Type != nil {
		return rawTypeKind(vs.Type)
	}
	if len(vs.Values) == 0 {
		return "int"
	}
	return exprValueKind(vs.Values[0])
}

func rawTypeKind(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "string"
		case "bool":
			return "bool"
		case "byte", "rune", "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
			return "int"
		}
	case *ast.SelectorExpr:
		// typed alias like package_basic.PackageType — treat as int
		return "int"
	}
	return ""
}

func exprValueKind(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		switch e.Kind {
		case token.INT:
			return "int"
		case token.STRING:
			return "string"
		case token.FLOAT:
			return "int"
		}
	case *ast.Ident:
		if e.Name == "true" || e.Name == "false" {
			return "bool"
		}
		if e.Name == "iota" {
			return "int"
		}
	case *ast.UnaryExpr, *ast.BinaryExpr:
		return "int"
	}
	return ""
}

func evalConstValue(expr ast.Expr, iotaVal int64, kind string) (string, bool) {
	switch kind {
	case "string":
		lit, ok := expr.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return "", false
		}
		return lit.Value, true
	case "bool":
		ident, ok := expr.(*ast.Ident)
		if !ok || (ident.Name != "true" && ident.Name != "false") {
			return "", false
		}
		return ident.Name, true
	default:
		v, ok := evalIntExpr(expr, iotaVal)
		if !ok {
			return "", false
		}
		return strconv.FormatInt(v, 10), true
	}
}

func evalIntExpr(expr ast.Expr, iotaVal int64) (int64, bool) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind != token.INT && e.Kind != token.FLOAT {
			return 0, false
		}
		v, err := strconv.ParseInt(e.Value, 0, 64)
		if err != nil {
			f, err := strconv.ParseFloat(e.Value, 64)
			if err != nil {
				return 0, false
			}
			return int64(f), true
		}
		return v, true
	case *ast.Ident:
		if e.Name == "iota" {
			return iotaVal, true
		}
		return 0, false
	case *ast.UnaryExpr:
		if e.Op != token.SUB && e.Op != token.ADD {
			return 0, false
		}
		v, ok := evalIntExpr(e.X, iotaVal)
		if !ok {
			return 0, false
		}
		if e.Op == token.SUB {
			return -v, true
		}
		return v, true
	case *ast.BinaryExpr:
		l, ok := evalIntExpr(e.X, iotaVal)
		if !ok {
			return 0, false
		}
		r, ok := evalIntExpr(e.Y, iotaVal)
		if !ok {
			return 0, false
		}
		switch e.Op {
		case token.ADD:
			return l + r, true
		case token.SUB:
			return l - r, true
		case token.MUL:
			return l * r, true
		case token.QUO:
			if r == 0 {
				return 0, false
			}
			return l / r, true
		case token.REM:
			if r == 0 {
				return 0, false
			}
			return l % r, true
		case token.AND:
			return l & r, true
		case token.OR:
			return l | r, true
		case token.XOR:
			return l ^ r, true
		case token.SHL:
			return l << r, true
		case token.SHR:
			return l >> r, true
		}
	}
	return 0, false
}

func enumMemberInlineComment(group *ast.CommentGroup) string {
	if group == nil || len(group.List) == 0 {
		return ""
	}
	text := strings.TrimSpace(strings.TrimLeft(group.List[0].Text, "/ \t"))
	if text == "" || strings.HasPrefix(text, TagPrefix) {
		return ""
	}
	return text
}
