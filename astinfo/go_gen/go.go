package go_gen

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/astinfo"
	"github.com/wanjm/gos/basic"
)

// GoGen generates a Go schema package of srpc client interfaces from servlet APIs.
type GoGen struct{}

func NewGoGen() *GoGen {
	return &GoGen{}
}

func (g *GoGen) GenerateCode(mp *astinfo.MainProject) {
	outDir := basic.Cfg.Generation.GoPath
	if outDir == "" {
		return
	}
	if !filepath.IsAbs(outDir) {
		outDir = filepath.Join(basic.Argument.SourcePath, outDir)
	}
	var err error
	outDir, err = filepath.Abs(outDir)
	if err != nil {
		fmt.Printf("Failed to resolve go gen directory: %v\n", err)
		return
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Printf("Failed to create go gen directory: %v\n", err)
		return
	}

	pkg := &astbasic.PkgBasic{
		Name:     "schema",
		FilePath: outDir,
		ModPath:  path.Join(mp.CurrentProject.ModPath, "schema"),
	}

	var allServices []*astinfo.Struct
	for _, pkgName := range mp.SortedPacakgeNames {
		pkgInfo := mp.Packages[pkgName]
		for _, structName := range pkgInfo.SortedStructNames {
			s := pkgInfo.Structs[structName]
			if len(s.MethodManager.Server) == 0 {
				continue
			}
			allServices = append(allServices, s)
		}
	}
	sort.Slice(allServices, func(i, j int) bool {
		return allServices[i].StructName < allServices[j].StructName
	})

	// Types/enums already written into the schema package (avoid redeclaration across files).
	writtenTypes := make(map[string]struct{})
	writtenEnums := make(map[string]struct{})

	for _, s := range allServices {
		g.genServiceFile(mp, pkg, s, writtenTypes, writtenEnums)
	}
}

func (g *GoGen) genServiceFile(mp *astinfo.MainProject, pkg *astbasic.PkgBasic, s *astinfo.Struct, writtenTypes, writtenEnums map[string]struct{}) {
	file := pkg.NewFile(astbasic.ToSnakeCase(s.StructName))
	file.GetImport(astbasic.SimplePackage("context", "context"))

	structs := make(map[string]*astinfo.Struct)
	aliases := make(map[string]*astinfo.Alias)
	enums := make(map[string]*astinfo.EnumDef)

	for _, m := range s.MethodManager.Server {
		for _, p := range m.Params {
			g.collectTypes(p.Type, mp, structs, aliases, enums)
		}
		for _, r := range m.Results {
			g.collectTypes(r.Type, mp, structs, aliases, enums)
		}
	}

	var sb strings.Builder

	// Enums first (only those not yet written).
	enumNames := sortedKeys(enums)
	for _, name := range enumNames {
		if _, ok := writtenEnums[name]; ok {
			continue
		}
		writtenEnums[name] = struct{}{}
		sb.WriteString(genEnumGo(enums[name]))
		sb.WriteString("\n")
	}

	// Structs and aliases (only those not yet written).
	typeNames := make([]string, 0, len(structs)+len(aliases))
	for name := range structs {
		if _, ok := writtenTypes[name]; !ok {
			typeNames = append(typeNames, name)
		}
	}
	for name := range aliases {
		if _, ok := writtenTypes[name]; !ok {
			typeNames = append(typeNames, name)
		}
	}
	sort.Strings(typeNames)
	for _, name := range typeNames {
		writtenTypes[name] = struct{}{}
		if st, ok := structs[name]; ok {
			sb.WriteString(g.genStruct(st, file))
			sb.WriteString("\n")
			continue
		}
		if al, ok := aliases[name]; ok {
			sb.WriteString(g.genAlias(al, file))
			sb.WriteString("\n")
		}
	}

	ifaceName := s.StructName + "ClientInterface"
	hostFn := "Get" + s.StructName + "Host"
	clientVar := s.StructName + "Client"
	hostVar := s.StructName + "Host"

	fmt.Fprintf(&sb, "// @gos type=srpc; host=%s();\n", hostFn)
	fmt.Fprintf(&sb, "type %s interface {\n", ifaceName)

	methods := make([]*astinfo.Method, len(s.MethodManager.Server))
	copy(methods, s.MethodManager.Server)
	sort.Slice(methods, func(i, j int) bool {
		return methods[i].Name < methods[j].Name
	})

	for _, m := range methods {
		url := m.Comment.Url
		fullURL := strings.ReplaceAll(filepath.Join(s.Comment.Url, url), "\\", "/")
		if m.Comment.Title != "" {
			fmt.Fprintf(&sb, "\t// %s\n", m.Comment.Title)
		}
		fmt.Fprintf(&sb, "\t// @gos url=%q\n", fullURL)
		sb.WriteString("\t")
		sb.WriteString(m.Name)
		sb.WriteString("(")
		sb.WriteString(g.genMethodParams(m, file))
		sb.WriteString(") (")
		sb.WriteString(g.genMethodResults(m, file))
		sb.WriteString(")\n")
	}
	sb.WriteString("}\n\n")

	fmt.Fprintf(&sb, "var %s string\n\n", hostVar)
	fmt.Fprintf(&sb, "func %s() string {\n\treturn %s\n}\n\n", hostFn, hostVar)
	fmt.Fprintf(&sb, "var %s %s\n", clientVar, ifaceName)

	file.AddBuilder(&sb)
	file.Save()
}

func (g *GoGen) genMethodParams(m *astinfo.Method, file *astinfo.GenedFile) string {
	var parts []string
	for i, p := range m.Params {
		name := p.Name
		if name == "" {
			if i == 0 {
				name = "ctx"
			} else {
				name = fmt.Sprintf("arg%d", i)
			}
		}
		parts = append(parts, name+" "+g.typeRef(p.Type, file))
	}
	return strings.Join(parts, ", ")
}

func (g *GoGen) genMethodResults(m *astinfo.Method, file *astinfo.GenedFile) string {
	var parts []string
	for _, r := range m.Results {
		parts = append(parts, g.typeRef(r.Type, file))
	}
	return strings.Join(parts, ", ")
}

func (g *GoGen) genStruct(s *astinfo.Struct, file *astinfo.GenedFile) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "type %s struct {\n", s.StructName)
	for _, field := range s.Fields {
		if field.Type == nil {
			continue
		}
		// Skip fields not serialized to API clients (json:"-").
		if field.GetJsonName() == "-" {
			continue
		}
		if field.Comment.CommentText != "" {
			fmt.Fprintf(&sb, "\t// %s\n", field.Comment.CommentText)
		}
		typeStr := g.typeRef(field.Type, file)
		tag := field.OriginalTag()
		if field.Name == "" {
			// Embedded field.
			if tag != "" {
				fmt.Fprintf(&sb, "\t%s %s\n", typeStr, tag)
			} else {
				fmt.Fprintf(&sb, "\t%s\n", typeStr)
			}
			continue
		}
		if tag != "" {
			fmt.Fprintf(&sb, "\t%s %s %s\n", field.Name, typeStr, tag)
		} else {
			fmt.Fprintf(&sb, "\t%s %s\n", field.Name, typeStr)
		}
	}
	sb.WriteString("}\n")
	return sb.String()
}

func (g *GoGen) genAlias(a *astinfo.Alias, file *astinfo.GenedFile) string {
	eq := " "
	if a.Equal {
		eq = " = "
	}
	return fmt.Sprintf("type %s%s%s\n", a.Name, eq, g.typeRef(a.Typer, file))
}

func (g *GoGen) typeRef(t astinfo.Typer, file *astinfo.GenedFile) string {
	if t == nil {
		return "any"
	}
	switch v := t.(type) {
	case *astinfo.PointerType:
		return "*" + g.typeRef(v.Typer, file)
	case *astinfo.ArrayType:
		return "[]" + g.typeRef(v.Typer, file)
	case *astinfo.MapType:
		return "map[" + g.typeRef(v.KeyTyper, file) + "]" + g.typeRef(v.ValueTyper, file)
	case *astinfo.Struct:
		if g.shouldCopyType(v.GoSource.Pkg.ModPath) {
			return v.StructName
		}
		return v.RefName(file)
	case *astinfo.Alias:
		if g.shouldCopyType(v.Gosourse.Pkg.ModPath) {
			return v.Name
		}
		return v.RefName(file)
	default:
		return t.RefName(file)
	}
}

func (g *GoGen) shouldCopyType(modPath string) bool {
	projectMod := astinfo.GlobalProject.CurrentProject.ModPath
	if projectMod == "" {
		return false
	}
	return modPath == projectMod || strings.HasPrefix(modPath, projectMod+"/")
}

func (g *GoGen) collectTypes(t astinfo.Typer, mp *astinfo.MainProject, structs map[string]*astinfo.Struct, aliases map[string]*astinfo.Alias, enums map[string]*astinfo.EnumDef) {
	if t == nil {
		return
	}
	switch v := t.(type) {
	case *astinfo.PointerType:
		g.collectTypes(v.Typer, mp, structs, aliases, enums)
	case *astinfo.ArrayType:
		g.collectTypes(v.Typer, mp, structs, aliases, enums)
	case *astinfo.MapType:
		g.collectTypes(v.KeyTyper, mp, structs, aliases, enums)
		g.collectTypes(v.ValueTyper, mp, structs, aliases, enums)
	case *astinfo.Alias:
		if !g.shouldCopyType(v.Gosourse.Pkg.ModPath) {
			return
		}
		if _, ok := aliases[v.Name]; ok {
			return
		}
		aliases[v.Name] = v
		g.collectTypes(v.Typer, mp, structs, aliases, enums)
	case *astinfo.Struct:
		if v.StructName == "Time" && v.GoSource.Pkg.ModPath == "time" {
			return
		}
		if !g.shouldCopyType(v.GoSource.Pkg.ModPath) {
			return
		}
		if _, ok := structs[v.StructName]; ok {
			return
		}
		structs[v.StructName] = v
		collectReferencedEnums(v, mp, enums)
		for _, field := range v.Fields {
			if field.Tags != nil && field.Tags[astinfo.JSON] == "-" {
				continue
			}
			g.collectTypes(field.Type, mp, structs, aliases, enums)
		}
	}
}

func collectReferencedEnums(s *astinfo.Struct, mp *astinfo.MainProject, seen map[string]*astinfo.EnumDef) {
	for _, field := range s.FlatFields() {
		if field.Comment.EnumName == "" {
			continue
		}
		enumDef := mp.LookupEnum(field.Comment.EnumName)
		if enumDef == nil {
			fmt.Printf("WARNING: enum %q not found for field %s::%s\n", field.Comment.EnumName, s.StructName, field.Name)
			continue
		}
		if _, ok := seen[enumDef.Name]; !ok {
			seen[enumDef.Name] = enumDef
		}
	}
}

func genEnumGo(enum *astinfo.EnumDef) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "// @gos enum=%s\n", enum.Name)
	sb.WriteString("const (\n")
	for _, m := range enum.Members {
		comment := m.Comment
		if m.DisplayWord != "" && comment == "" {
			comment = m.DisplayWord
		}
		if comment != "" {
			fmt.Fprintf(&sb, "\t%s = %s // %s\n", m.GoName, m.Value, comment)
		} else {
			fmt.Fprintf(&sb, "\t%s = %s\n", m.GoName, m.Value)
		}
	}
	sb.WriteString(")\n")
	return sb.String()
}

func sortedKeys[V any](m map[string]V) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
