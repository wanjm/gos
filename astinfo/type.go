package astinfo

// 后续考虑建一个Typer的map，这样所有相同的Typer在内存中就一个对象，便于层次比较；
// 统一的工作需要在Package.ParseType函数中完成;
type Typer interface {
	// 使用genFile作为参数的目的是，在生成代码时，需要根据genFile来生成import代码；
	// 配合GenFileName，该函数完成是在GenFileName引用该Type时，变量类型的全称；
	// genFile 为空，则简单返回类型名字；
	// genFile不为空，pkg跟结构体相同，则返回结构体名字；
	// genFile不为空，pkg跟结构体不同，则返回pkg.Name + "." + 结构体名字；同时向genFile添加import代码；
	RefName(genFile *GenedFile) string
	// 该变量的全名pkgModPath.StructName达到全局唯一的目的
	IDName() string
	Constructor
}
type Constructor interface {
	GenConstructCode(genFile *GenedFile, wire bool) string
}

func GetConstructor(typer Typer) Constructor {
	if constructor, ok := typer.(Constructor); ok {
		return constructor
	}
	if IsPointer(typer) {
		return GetConstructor(typer.(*PointerType).Typer)
	}
	return nil
}

func GetBasicType(typer Typer) Typer {
	if t, ok := typer.(*PointerType); ok {
		return GetBasicType(t.Typer)
	}
	return typer
}

func IsPointer(typer Typer) bool {
	_, ok := typer.(*PointerType)
	return ok
}

// isRawType
func IsRawType(typer Typer) bool {
	_, ok := typer.(*RawType)
	return ok
}

func PointerDepth(typer Typer) int {
	if !IsPointer(typer) {
		return 0
	}
	return typer.(*PointerType).Depth
}

// 解析原生类型，主要是生成swagger要用；
type BaseType struct {
	typeName string
}

// func (b *BaseType) IsPointer() bool {
// 	return false
// }

func (b *BaseType) RefName(_ *GenedFile) string {
	return b.typeName
}

func (b *BaseType) IDName() string {
	return b.typeName
}

func (b *BaseType) GenConstructCode(genFile *GenedFile, _ bool) string {
	switch b.typeName {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "0"
	case "float32", "float64":
		return "0.0"
	case "string":
		return `""`
	case "bool":
		return "false"
	}
	return b.typeName
}

type ArrayType struct {
	Typer
}

// RefName
func (a *ArrayType) RefName(genFile *GenedFile) string {
	return "[]" + a.Typer.RefName(genFile)
}

type MapType struct {
	BaseType
	KeyTyper   Typer
	ValueTyper Typer
}

type RawType struct {
	BaseType
}

type PointerType struct {
	Typer
	Depth int
}

//	func (p *PinterType) IsPointer() bool {
//		return true
//	}
//
// NewPointerType
func NewPointerType(typer Typer) *PointerType {
	depth := 1
	if ptr, ok := typer.(*PointerType); ok {
		depth = ptr.Depth + 1
	}
	return &PointerType{
		Typer: typer,
		Depth: depth,
	}
}

func (p *PointerType) RefName(genFile *GenedFile) string {
	return "*" + p.Typer.RefName(genFile)
}

func (p *PointerType) GenConstructCode(genFile *GenedFile, wire bool) string {
	var code = p.Typer.GenConstructCode(genFile, wire)
	if IsPointer(p.Typer) {
		return "getAddr(" + code + ",)"
	} else {
		return "&" + code
	}
}

var rawTypeMap = map[string]*RawType{}

func init() {
	rawType := []string{"string", "bool", "byte", "rune",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"any", "error",
		"uintptr", "complex128", "complex64",
	}
	for _, t := range rawType {
		rawTypeMap[t] = &RawType{
			BaseType: BaseType{
				typeName: t,
			},
		}
	}
}

func GetRawType(name string) *RawType {
	return rawTypeMap[name]
}
