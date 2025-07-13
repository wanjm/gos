package astinfo

type Typer interface {
	IsPointer() bool
	Name(genFile *GenedFile) string
}

// 解析原生类型，主要是生成swagger要用；
type BaseType struct {
	typeName string
}

func (b *BaseType) IsPointer() bool {
	return false
}

func (b *BaseType) Name(_ *GenedFile) string {
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
}

func (p *PinterType) IsPointer() bool {
	return true
}

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
