package astinfo

type Typer interface {
	IsPointer() bool
	Name() string
}

// 解析原生类型，主要是生成swagger要用；
type BaseType struct {
	isPointer bool
	typeName  string
}

func (b *BaseType) IsPointer() bool {
	return b.isPointer
}

func (b *BaseType) Name() string {
	return b.typeName
}

type ArrayType struct {
	BaseType
}

type RawType struct {
	BaseType
}

var rawTypeMap = map[string]*RawType{}

func init() {
	rawType := []string{"int", "string", "bool", "float64", "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "float32"}
	for _, t := range rawType {
		rawTypeMap[t] = &RawType{
			BaseType: BaseType{
				isPointer: false,
				typeName:  t,
			},
		}
	}
}

func GetRawType(name string) *RawType {
	return rawTypeMap[name]
}
