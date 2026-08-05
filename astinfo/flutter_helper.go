package astinfo

// FlutterHelper generates Dart type/default/parse snippets for a Go Typer.
// Methods mirror FlutterGen.mapType, genTypeDefault, and genTypeParse.
type FlutterHelper interface {
	GenDartType() string
	GenDefaultValue() string
	GenTypeParse(expr string) string
}

func (b *BaseType) GenDartType() string {
	return "dynamic"
}

func (b *BaseType) GenDefaultValue() string {
	return "null"
}

func (b *BaseType) GenTypeParse(expr string) string {
	return expr
}

func (m *MissingType) GenDartType() string {
	return "dynamic"
}

func (m *MissingType) GenDefaultValue() string {
	return "null"
}

func (m *MissingType) GenTypeParse(expr string) string {
	return expr
}

func (r *RawType) GenDartType() string {
	switch r.IDName() {
	case "string":
		return "String"
	case "byte", "int", "int8", "int16", "int32", "int64", "uint", "uint32", "uint64", "uint8", "uint16":
		return "int"
	case "float32", "float64":
		return "double"
	case "bool":
		return "bool"
	default:
		return "dynamic"
	}
}

func (r *RawType) GenDefaultValue() string {
	switch r.IDName() {
	case "string":
		return `""`
	case "int", "int8", "int16", "int32", "int64", "uint", "uint32", "uint64", "uint8", "uint16", "byte":
		return "0"
	case "float32", "float64":
		return "0.0"
	case "bool":
		return "false"
	default:
		return "null"
	}
}

func (r *RawType) GenTypeParse(expr string) string {
	switch r.IDName() {
	case "string":
		return expr + ` as String? ?? ""`
	case "int", "int8", "int16", "int32", "int64", "uint", "uint32", "uint64", "uint8", "uint16", "byte":
		return "(" + expr + " as num? ?? 0).toInt()"
	case "float32", "float64":
		return "(" + expr + " as num? ?? 0.0).toDouble()"
	case "bool":
		return expr + " ?? false"
	default:
		return expr
	}
}

func (a *ArrayType) GenDartType() string {
	elem := GetBasicType(a.Typer)
	if raw, ok := elem.(*RawType); ok {
		if id := raw.IDName(); id == "byte" || id == "uint8" {
			return "String"
		}
	}
	return "List<" + GetBasicType(a.Typer).GenDartType() + ">"
}

func (a *ArrayType) GenDefaultValue() string {
	elem := GetBasicType(a.Typer)
	if raw, ok := elem.(*RawType); ok {
		if id := raw.IDName(); id == "byte" || id == "uint8" {
			return `""`
		}
	}
	return "const []"
}

func (a *ArrayType) GenTypeParse(expr string) string {
	elem := GetBasicType(a.Typer)
	if raw, ok := elem.(*RawType); ok {
		if id := raw.IDName(); id == "byte" || id == "uint8" {
			return expr + `?.toString() ?? ""`
		}
	}
	elemParse := GetBasicType(a.Typer).GenTypeParse("e")
	return "(" + expr + " as List? ?? []).map((e) => " + elemParse + ").toList()"
}

func (p *PointerType) GenDartType() string {
	return GetBasicType(p.Typer).GenDartType()
}

func (p *PointerType) GenDefaultValue() string {
	return GetBasicType(p.Typer).GenDefaultValue()
}

func (p *PointerType) GenTypeParse(expr string) string {
	return GetBasicType(p.Typer).GenTypeParse(expr)
}

func (v *Struct) GenDartType() string {
	if v.StructName == "Time" && v.GoSource.Pkg.ModPath == "time" {
		return "DateTime"
	}
	return v.StructName
}

func (v *Struct) GenDefaultValue() string {
	if v.StructName == "Time" && v.GoSource.Pkg.ModPath == "time" {
		return "DateTime.fromMillisecondsSinceEpoch(0)"
	}
	return v.StructName + ".fromJson({})"
}

func (v *Struct) GenTypeParse(expr string) string {
	if v.StructName == "Time" && v.GoSource.Pkg.ModPath == "time" {
		return expr + " != null && " + expr + ".toString().isNotEmpty ? DateTime.parse(" + expr + ".toString()) : DateTime.fromMillisecondsSinceEpoch(0)"
	}
	return v.StructName + ".fromJson(" + expr + " ?? {})"
}

func (a *Alias) GenDartType() string {
	return GetBasicType(a.Typer).GenDartType()
}

func (a *Alias) GenDefaultValue() string {
	return GetBasicType(a.Typer).GenDefaultValue()
}

func (a *Alias) GenTypeParse(expr string) string {
	return GetBasicType(a.Typer).GenTypeParse(expr)
}

func (i *Interface) GenDartType() string {
	return "dynamic"
}

func (i *Interface) GenDefaultValue() string {
	return "null"
}

func (i *Interface) GenTypeParse(expr string) string {
	return expr
}
