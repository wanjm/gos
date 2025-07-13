package astinfo

type Typer interface {
	IsPointer() bool
	Name() string
}

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

// 解析原生类型，主要是生成swagger要用；
type ArrayType struct {
	BaseType
}
