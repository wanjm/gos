package astinfo

import (
	"go/ast"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/wanjm/gos/astbasic"
)

// @goservlet prpc=xxx; servlet=xxx; servlet; prpc
// @goservlet type=xxx ;  prpc, servlet, websocket, restful,
// @goservlet group=xxx; 如果不存在，则跟type同名
type structComment struct {
	GroupName  string
	serverType string // NONE, RpcStruct, ServletStruct·
	Url        string // 服务的url, 对所有的方法都有效
	AutoGen    bool
	TableName  string
	DbVarible  string
	class      *Struct
}

func (comment *structComment) dealValuePair(key, value string) {
	if value != "" {
		value = strings.Trim(value, "\"")
	}
	switch key {
	case Prpc:
		comment.AutoGen = true
		comment.serverType = Prpc
		if len(value) == 0 {
			comment.GroupName = Prpc
		} else {
			comment.GroupName = value
		}
	case Servlet:
		comment.AutoGen = true
		comment.serverType = Servlet
		if len(value) == 0 {
			comment.GroupName = Servlet
		} else {
			comment.GroupName = value
		}
	case Group:
		comment.GroupName = value
	case Type:
		comment.AutoGen = true
		comment.serverType = value
		if len(comment.GroupName) == 0 {
			comment.GroupName = comment.serverType
		}
	case Url:
		comment.Url = value
	case AutoGen:
		comment.AutoGen = true
	case tblName:
		if value == "" {
			value = astbasic.ToSnakeCase(comment.class.StructName)
		}
		comment.TableName = value
		if comment.DbVarible == "" {
			comment.DbVarible = "DB"
		}
	case dbVarible:
		comment.DbVarible = astbasic.Capitalize(value)
	}
}

// Struct 表示一个Go结构体的基本信息
type Struct struct {
	StructName string    // 结构体名称
	GoSource   *Gosourse //该变量在解析结构体时赋值，也就是说该变量不为空，则该结构体已经被解析；
	// Pkg        *Package  //该变量肯定不为空，但是goSource不一定；
	astRoot *ast.TypeSpec
	// genDecl       *ast.GenDecl
	// astStructType *ast.StructType
	Comment       structComment
	Fields        []*Field
	TypeParameter []*Field
	// FieldMap      map[string]*Field
	MethodManager
	// TODO: 后续添加字段和方法解析
	ref *spec.Ref
}

func (v *Struct) RefName(genFile *GenedFile) string {
	pkg := v.GoSource.Pkg
	if genFile == nil || pkg.IsSame(genFile.Pkg) {
		return v.StructName
	}
	impt := genFile.GetImport(&pkg.PkgBasic)
	return impt.Name + "." + v.StructName
}

func (v *Struct) IDName() string {
	return v.GoSource.Pkg.ModPath + "." + v.StructName
}

// 某些field需要wire，但是却没有名字，所以需要处理
func getWireField(field *Field) *Field {
	// 1. 原始类型，不需要wire；（可以通过default直接构造，或者make构造，或者不写，使用系统的默认0值）
	if IsGolangType(field.Type) {
		return nil
	}
	var name = field.wriedName()
	if name == "" {
		return nil
	}
	result := *field
	result.Name = name
	return &result
}

func (field *Field) wriedName() string {
	name := field.Name
	if name == "" {
		t := GetBasicType(field.Type)
		if t != nil {
			name = t.RefName(nil)
		}
	}
	if name == "" || (name[0] <= 'z' && name[0] >= 'a') {
		return ""
	}

	if field.Tags["wire"] == `-` {
		return ""
	}
	return name
}

// 这里有两种情况。如果定义一个统一的wire，需要考虑一下；
// 1. 自动注入，dal的情况；
// 2. request请求的自动创建；
// 3. 系统原始类型；
// 4. 嵌套的其他结构体；对于嵌套的结构体，在注入时，是必须要找到的；
// 5. 初步考虑可以将wire变量定义为必须注入内容结构体变量；
// wire为true表示必须绑定结构体等；
func (v *Struct) GenConstructCode(genFile *GenedFile, wire bool) string {
	result := genFile.GetImport(&v.GoSource.Pkg.PkgBasic)
	var sb strings.Builder
	if result.Name != "" {
		sb.WriteString(result.Name)
		sb.WriteString(".")
	}
	sb.WriteString(v.StructName + "{\n")
	//结尾不能有\n,否则后续代码不好写，有语法错误；如：最后两行会有语法错误
	// getAddr(Strurct{
	//}
	//)
	// 1. 有default，则wire；
	// 2. wire为ture，且不是简单结构体（needWire），则寻找值去绑定；
	for _, field := range v.Fields {
		var name = field.wriedName()
		if name != "" && wire {
			sb.WriteString(name + ":")
			//此处需要考虑default为字符串等各种情况；
			// RawType是原始数据类型；不包含map，chan；
			if rt, ok := field.Type.(*RawType); ok {
				v, ok := field.Tags["default"]
				if ok {
					if rt.typeName == "string" {
						v = `"` + v + `"`
					}
				}
				sb.WriteString(v)
			} else {
				sb.WriteString(field.GenVariableCode(genFile, wire))
			}
			sb.WriteString(",\n")
		}

	}
	sb.WriteString("}")

	return sb.String()
}

// RequiredFields 返回结构体需要检查依赖注入的field；
// 1. 原始类型不要依赖注入；(天然跳过有依赖注入的原始类型)
// 2. 有wireName的需要依赖注入;
func (v *Struct) RequiredFields() []*Field {
	var requiredFields []*Field
	for _, field := range v.Fields {
		// 过滤原始类型
		if IsGolangType(field.Type) {
			continue
		}
		if wireName := field.wriedName(); wireName != "" {
			wireField := *field
			wireField.Name = wireName
			requiredFields = append(requiredFields, &wireField)
		}
	}
	return requiredFields
}

// GeneredFields 返回结构体自己
func (v *Struct) GeneredFields() []*Field {
	// 创建一个表示结构体自身的变量，且是指针格式；
	field := NewSimpleField(NewPointerType(v), "")
	return []*Field{field}
}
func (field *Struct) GenNilCode(file *GenedFile) string {
	var sb strings.Builder
	for _, f := range field.Fields {
		switch f.Type.(type) {
		case *ArrayType, *Struct:
			sb.WriteString("{\na:=&a." + f.Name + "\n")
			sb.WriteString(f.GenNilCode(file))
			sb.WriteString("}\n")
		case *PointerType:
			sb.WriteString("{\na:=a." + f.Name + "\n")
			sb.WriteString(f.GenNilCode(file))
			sb.WriteString("}\n")
		}
	}
	return sb.String()
}

// GenerateDependcyCode 生成创建结构体对象的代码
func (v *Struct) GenerateDependcyCode(goGenerated *GenedFile) string {
	a := v.GeneredFields()[0]
	return a.Type.GenConstructCode(goGenerated, true)
}
func (v *Struct) GetInfo() string {
	return "struct " + v.StructName + " in " + v.GoSource.Path
}

// 不一定每次newStruct时都会有goSrouce，所以此时只能传Pkg；
// func NewStruct(name string, pkg *Package) *Struct {
// 	class := &Struct{
// 		StructName: name,
// 		Pkg:        pkg,
// 	}
// 	pkg.Structs[name] = class
// 	pkg.Types[name] = class
// 	return class
// }

func NewStruct(goSource *Gosourse, astRoot *ast.TypeSpec) *Struct {
	name := astRoot.Name.Name
	iface := &Struct{
		StructName: name,
		GoSource:   goSource,
		astRoot:    astRoot,
	}
	iface.Comment.class = iface
	pkg := goSource.Pkg
	pkg.Structs[name] = iface
	pkg.Types[name] = iface
	return iface
}

// initGenDecl
// func (v *Struct) initGenDecl(genDecl *ast.GenDecl, structType *ast.StructType) {
// 	v.genDecl = genDecl
// 	v.astStructType = structType
// }

// 解析结构体的注释和字段
func (v *Struct) Parse() error {
	if err := v.ParseComment(); err != nil {
		return err
	}
	if v.astRoot.TypeParams != nil {
		v.TypeParameter = parseFields(v.astRoot.TypeParams.List, v.GoSource, nil)
	}
	return v.ParseField()
}

// parseComment
func (class *Struct) ParseComment() error {
	parseComment(class.astRoot.Doc, &class.Comment)
	return nil
}

// parseField
func (v *Struct) ParseField() error {
	// v.goSource在解析结构体时，被赋值，解析field也是在解析结构体时，所以v.goSource不为空
	v.Fields = parseFields(v.astRoot.Type.(*ast.StructType).Fields.List, v.GoSource, FieldListToMap(v.TypeParameter))
	// v.FieldMap = make(map[string]*Field)
	// for _, field := range v.Fields {
	// 	v.FieldMap[field.Name] = field
	// }
	return nil
}
