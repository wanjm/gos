package astinfo

// 后续考虑建一个Typer的map，这样所有相同的Typer在内存中就一个对象，便于层次比较；
// 统一的工作需要在Package.ParseType函数中完成;
type Typer interface {
	// 使用genFile作为参数的目的是，在生成代码时，需要根据genFile来生成import代码；
	// 配合GenFileName，该函数完成是在GenFileName引用该Type时，变量类型的全称；
	Name(genFile *GenedFile) string
	// 该变量的全名（目前时pkg.StructName）,后续可能需要改为pkgModPath.StructName达到全局唯一的目的
	// 保证该类的全局唯一是本函数的目的；
	FullName() string
}

func IsPointer(typer Typer) bool {
	_, ok := typer.(*PinterType)
	return ok
}
func PointerDepth(typer Typer) int {
	if !IsPointer(typer) {
		return 0
	}
	return typer.(*PinterType).Depth
}

// 解析原生类型，主要是生成swagger要用；
type BaseType struct {
	typeName string
}

// func (b *BaseType) IsPointer() bool {
// 	return false
// }

func (b *BaseType) Name(_ *GenedFile) string {
	return b.typeName
}

func (b *BaseType) FullName() string {
	return b.typeName
}

type ArrayType struct {
	Typer
}

// Name
func (a *ArrayType) Name(genFile *GenedFile) string {
	return "[]" + a.Typer.Name(genFile)
}

type RawType struct {
	BaseType
}

type PinterType struct {
	Typer
	Depth int
}

// func (p *PinterType) IsPointer() bool {
// 	return true
// }

func (p *PinterType) Name(genFile *GenedFile) string {
	return "*" + p.Typer.Name(genFile)
}

var rawTypeMap = map[string]*RawType{}

func init() {
	rawType := []string{"int", "string", "bool", "float64", "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "float32"}
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
