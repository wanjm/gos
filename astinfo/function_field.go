package astinfo

import "go/ast"

type FunctionField struct {
	GoSource *Gosourse
	Name     string //函数名

	Params  []*Field // method params, 下标0是request
	Results []*Field // method results（output)	Params      []*Field // method params, 下标0是request

}

func (f *FunctionField) RequiredFields() []*Field {
	return f.Params
}

// GeneredFields 返回函数生成的字段
func (f *FunctionField) GeneredFields() []*Field {
	return f.Results
}

func (f *FunctionField) parseParameter(paramType *ast.FuncType) bool {
	//Params参数不可能为nil
	f.Params = parseFields(paramType.Params.List, f.GoSource, nil)
	//Results返回值可能为nil
	if paramType.Results != nil {
		f.Results = parseFields(paramType.Results.List, f.GoSource, nil)
	}
	return true
}

// 从ast.Field中解析出参数
func parseFields(params []*ast.Field, goSource *Gosourse, typeMap map[string]*Field) []*Field {
	var result []*Field
	for _, param := range params {
		field := NewField(param, goSource)
		field.Parse(typeMap)
		if len(param.Names) != 0 {
			for _, name := range param.Names {
				field1 := *field
				field1.Name = name.Name
				result = append(result, &field1)
			}
		} else {
			//没有参数名，基本不会出现
			result = append(result, field)
		}
	}
	return result
}
