package astinfo

import (
	"fmt"
	"go/ast"
	"sort"
	"strings"
)

type EntityGen struct {
	project *MainProject
}

func NewEntityGen(project *MainProject) *EntityGen {
	return &EntityGen{
		project: project,
	}
}

func (eg *EntityGen) GenerateCode() {
	// Group schema structs by package
	pkgSchemaMap := make(map[*Package][]*Struct)
	for _, pkgName := range eg.project.SortedPacakgeNames {
		pkg := eg.project.Packages[pkgName]
		var schemaStructs []*Struct
		for _, structName := range pkg.SortedStructNames {
			s := pkg.Structs[structName]
			if s.Comment.EntityName != "" {
				schemaStructs = append(schemaStructs, s)
			}
		}
		if len(schemaStructs) > 0 {
			pkgSchemaMap[pkg] = schemaStructs
		}
	}

	// Generate code per package in a single schema.gen.go file
	for pkg, schemaStructs := range pkgSchemaMap {
		genFile := pkg.NewFile("schema")

		// Generate FromEntity methods for all schema structs in this package
		for _, schemaStruct := range schemaStructs {
			eg.generateFromEntityToFile(schemaStruct, genFile)
		}

		// Generate FromEntitys methods for array types in this package
		eg.generateFromEntitysForPackage(pkg, genFile)

		genFile.Save()
	}
}

func (eg *EntityGen) generateFromEntityToFile(schemaStruct *Struct, genFile *GenedFile) {
	entityName := schemaStruct.Comment.EntityName
	if entityName == "" {
		return
	}

	// Find entity struct
	entityStruct := eg.findEntityStruct(schemaStruct, entityName)
	if entityStruct == nil {
		fmt.Printf("WARNING: entity struct %s not found for schema %s.%s\n", entityName, schemaStruct.GoSource.Pkg.ModPath, schemaStruct.StructName)
		return
	}

	// Generate FromEntity method
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("func (s *%s) FromEntity(e *%s) {\n", schemaStruct.StructName, entityStruct.RefName(genFile)))

	// Get entity import
	entityImport := genFile.GetImport(&entityStruct.GoSource.Pkg.PkgBasic)
	if entityImport.Name != "" {
		// Entity is from different package, use import alias
		sb.WriteString("\tif e == nil {\n\t\treturn\n\t}\n")
	}

	// Match fields by name and copy
	entityFieldMap := make(map[string]*Field)
	for _, field := range entityStruct.Fields {
		entityFieldMap[field.Name] = field
	}

	for _, schemaField := range schemaStruct.Fields {
		entityField, exists := entityFieldMap[schemaField.Name]
		if !exists {
			continue // Skip fields that don't exist in entity
		}

		// Validate types match
		schemaType := GetBasicType(schemaField.Type)
		entityType := GetBasicType(entityField.Type)

		// Compare types
		if !eg.typesMatch(schemaType, entityType) {
			fmt.Printf("ERROR: type mismatch for field %s in %s.%s: schema has %s, entity has %s\n",
				schemaField.Name, schemaStruct.GoSource.Pkg.ModPath, schemaStruct.StructName,
				schemaType.IDName(), entityType.IDName())
			continue
		}

		// Generate field assignment
		sb.WriteString(fmt.Sprintf("\ts.%s = e.%s\n", schemaField.Name, entityField.Name))
	}

	// Check if FormatEntity method exists
	if eg.hasFormatEntityMethod(schemaStruct) {
		sb.WriteString("\ts.FormatEntity()\n")
	}

	sb.WriteString("}\n\n")
	genFile.AddBuilder(&sb)
}

func (eg *EntityGen) findEntityStruct(schemaStruct *Struct, entityName string) *Struct {
	// Search entity map first
	if entityStruct, exists := eg.project.EntityMap[entityName]; exists {
		return entityStruct
	}
	return nil
}

func (eg *EntityGen) typesMatch(schemaType, entityType Typer) bool {
	// Compare by IDName for exact match
	if schemaType.IDName() == entityType.IDName() {
		return true
	}

	// Handle pointer types - compare underlying types
	if schemaPtr, ok := schemaType.(*PointerType); ok {
		if entityPtr, ok := entityType.(*PointerType); ok {
			return eg.typesMatch(schemaPtr.Typer, entityPtr.Typer)
		}
		// Schema is pointer but entity is not - might be okay for assignment
		return eg.typesMatch(schemaPtr.Typer, entityType)
	}

	// Handle aliases
	if schemaAlias, ok := schemaType.(*Alias); ok {
		return eg.typesMatch(schemaAlias.Typer, entityType)
	}
	if entityAlias, ok := entityType.(*Alias); ok {
		return eg.typesMatch(schemaType, entityAlias.Typer)
	}

	// For raw types, compare by name
	if schemaRaw, ok := schemaType.(*RawType); ok {
		if entityRaw, ok := entityType.(*RawType); ok {
			return schemaRaw.typeName == entityRaw.typeName
		}
	}

	return false
}

func (eg *EntityGen) hasFormatEntityMethod(schemaStruct *Struct) bool {
	// Check all methods in the package that have this struct as receiver
	pkg := schemaStruct.GoSource.Pkg
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 && funcDecl.Name.Name == "FormatEntity" {
					recvType := funcDecl.Recv.List[0].Type
					var recvName string
					switch rt := recvType.(type) {
					case *ast.Ident:
						recvName = rt.Name
					case *ast.StarExpr:
						if ident, ok := rt.X.(*ast.Ident); ok {
							recvName = ident.Name
						}
					}
					if recvName == schemaStruct.StructName {
						return true
					}
				}
			}
		}
	}
	return false
}

func (eg *EntityGen) generateFromEntitysForPackage(pkg *Package, genFile *GenedFile) {
	// Find all array type aliases where element type has entity annotation
	keys := make([]string, 0, len(pkg.Types))
	for k := range pkg.Types {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		typer := pkg.Types[name]
		if typer == nil {
			continue
		}

		// Check if it's an alias to an array type
		alias, ok := typer.(*Alias)
		if !ok {
			continue
		}

		// Check if underlying type is an array
		arrayType, ok := alias.Typer.(*ArrayType)
		if !ok {
			continue
		}

		// Check if element type is a struct with entity annotation
		// Handle pointer types: []*StructName
		elemType := arrayType.Typer
		var elemStruct *Struct

		// Check if it's a pointer type
		if ptrType, isPtr := elemType.(*PointerType); isPtr {
			elemType = GetBasicType(ptrType.Typer)
		} else {
			elemType = GetBasicType(elemType)
		}

		var isStruct bool
		elemStruct, isStruct = elemType.(*Struct)
		if !isStruct {
			continue
		}

		if elemStruct.Comment.EntityName == "" {
			continue
		}

		// Generate FromEntitys method to the same file
		eg.generateFromEntitysForArrayToFile(alias, elemStruct, arrayType.Typer, genFile)
	}
}

func (eg *EntityGen) generateFromEntitysForArrayToFile(arrayAlias *Alias, elemStruct *Struct, elemType Typer, genFile *GenedFile) {
	entityName := elemStruct.Comment.EntityName
	entityStruct := eg.findEntityStruct(elemStruct, entityName)
	if entityStruct == nil {
		fmt.Printf("WARNING: entity struct %s not found for array type %s.%s\n", entityName, arrayAlias.Gosourse.Pkg.ModPath, arrayAlias.Name)
		return
	}

	// Get entity import
	genFile.GetImport(&entityStruct.GoSource.Pkg.PkgBasic)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("func (l *%s) FromEntitys(entitys []*%s) {\n", arrayAlias.Name, entityStruct.RefName(genFile)))
	sb.WriteString("\tfor _, e := range entitys {\n")

	// Check if element type is a pointer
	if _, isPtr := elemType.(*PointerType); isPtr {
		sb.WriteString(fmt.Sprintf("\t\titem := &%s{}\n", elemStruct.StructName))
	} else {
		sb.WriteString(fmt.Sprintf("\t\titem := %s{}\n", elemStruct.StructName))
	}
	sb.WriteString("\t\titem.FromEntity(e)\n")

	if _, isPtr := elemType.(*PointerType); isPtr {
		sb.WriteString("\t\t*l = append(*l, item)\n")
	} else {
		sb.WriteString("\t\t*l = append(*l, &item)\n")
	}
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	genFile.AddBuilder(&sb)
}
