package astinfo

type Parser interface {
	Parse() error
}

type CodeGenerator interface {
	// 需要GoGenerated，作为参数的主要原因是，生成过程中需要向代码中添加import；

	Generate(goGenerated *GenedFile) error
}

type Import struct {
	Name string
	Path string
}
