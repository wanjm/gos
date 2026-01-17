package astinfo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/wanjm/gos/basic"
)

type SchemaType interface {
	InitSchema(*spec.Schema, *Swagger)
}

func (r *RawType) InitSchema(schema *spec.Schema, swagger *Swagger) {
	// 获取原始类型对应到swagger的类型
	var name = "integer"
	switch r.typeName {
	case "string":
		name = "string"
	case "array":
		name = "array"
	case "map":
		name = "object"
	case "bool":
		name = "bool"
	case "float32", "float64":
		name = "number"
	}
	schema.Type = []string{name}
}

func (r *ArrayType) InitSchema(schema *spec.Schema, swagger *Swagger) {
	schema.Type = []string{"array"}
	basicType := GetBasicType(r.Typer)
	if raw, ok := basicType.(*RawType); ok {
		if raw.typeName == "byte" {
			schema.Type = []string{"string"}
			return
		}
	}
	schema.Items = &spec.SchemaOrArray{
		Schema: &spec.Schema{},
	}
	basicType.(SchemaType).InitSchema(schema.Items.Schema, swagger)
}

// func (m *MapType) InitSchema(schema *spec.Schema, swagger *Swagger) {
// 	schema.Type = []string{"object"}
// 	schema.AdditionalProperties = &spec.SchemaOrBool{
// 		Schema: &spec.Schema{},
// 	}
// }

func (s *Struct) InitSchema(schema *spec.Schema, swagger *Swagger) {
	// schema.Ref = spec.Ref{
	if s.ref == nil {
		swagger.getStructRef(s)
	}
	schema.Ref = *s.ref
}
func (s *Alias) InitSchema(schema *spec.Schema, swagger *Swagger) {
	// schema.Ref = spec.Ref{
	basicType := GetBasicType(s.Typer)
	basicType.(SchemaType).InitSchema(schema, swagger)
}

func (e *BaseType) InitSchema(schema *spec.Schema, swagger *Swagger) {
}
func (p *PointerType) InitSchema(schema *spec.Schema, swagger *Swagger) {
	basicType := GetBasicType(p.Typer)
	basicType.(SchemaType).InitSchema(schema, swagger)
}

type Swagger struct {
	swag    *spec.Swagger
	project *MainProject
	// definitions    map[*Struct]*spec.Ref
	responseResult *Struct
}

func NewSwagger(project *MainProject) (result *Swagger) {
	var swag = &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Contact: &spec.ContactInfo{},
					License: nil,
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{},
				},
			},
			Paths: &spec.Paths{
				Paths: make(map[string]spec.PathItem),
				VendorExtensible: spec.VendorExtensible{
					Extensions: nil,
				},
			},
			Definitions:         make(map[string]spec.Schema),
			SecurityDefinitions: make(map[string]*spec.SecurityScheme),
		},
		VendorExtensible: spec.VendorExtensible{
			Extensions: nil,
		},
	}
	result = &Swagger{
		swag:    swag,
		project: project,
		// definitions: make(map[string]*spec.Ref),
	}

	if len(project.Cfg.SwaggerCfg.UrlPrefix) > 0 {
		if !strings.HasPrefix(project.Cfg.SwaggerCfg.UrlPrefix, "/") {
			project.Cfg.SwaggerCfg.UrlPrefix = "/" + project.Cfg.SwaggerCfg.UrlPrefix
		}
		project.Cfg.SwaggerCfg.UrlPrefix = strings.TrimSuffix(project.Cfg.SwaggerCfg.UrlPrefix, "/")
	}
	result.initResponseResult()
	return
}

func initOperation(title string) *spec.Operation {
	return &spec.Operation{
		OperationProps: spec.OperationProps{
			ID:           "",
			Description:  "",
			Summary:      title,
			Security:     nil,
			ExternalDocs: nil,
			Deprecated:   false,
			Tags:         []string{},
			Consumes:     []string{},
			Produces:     []string{},
			Schemes:      []string{},
			Parameters:   []spec.Parameter{},
			Responses: &spec.Responses{
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{},
				},
				ResponsesProps: spec.ResponsesProps{
					Default:             nil,
					StatusCodeResponses: make(map[int]spec.Response),
				},
			},
		},
		VendorExtensible: spec.VendorExtensible{
			Extensions: spec.Extensions{},
		},
	}
}
func (swagger *Swagger) addServletFromFunctionManager(pkg *MethodManager) {
	paths := swagger.swag.Paths.Paths
	for _, servlet := range pkg.Server {
		comment := servlet.Comment
		var url = strings.Trim(comment.Url, "\"")
		if len(url) == 0 {
			fmt.Printf("servlet %s has no url\n", servlet.Name)
			continue
		}
		pathItem := spec.PathItem{}
		operation := initOperation(comment.title)
		var parameter []spec.Parameter
		switch comment.Method {
		case POST, "":
			pathItem.Post = operation
			var props spec.SchemaProps
			_ = props
			if len(servlet.Params) > 1 && servlet.Params[1].Type != nil {
				t := GetBasicType(servlet.Params[1].Type)
				ref := swagger.getStructRef(t.(*Struct))
				parameter = append(parameter, spec.Parameter{
					ParamProps: spec.ParamProps{
						Name:     "body",
						In:       "body",
						Required: true,
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Ref: *ref,
							},
						},
					},
				})
			}

		case GET:
			pathItem.Get = operation
		default:
			fmt.Printf("servlet %s has invalid method %s,which is not supported\n", servlet.Name, servlet.Comment.Method)
			continue
		}
		operation.Parameters = parameter
		var objFieldPtr *Field
		if len(servlet.Results) > 1 {
			field0 := servlet.Results[0]
			// if field0.Type == nil {
			// 	field0.findStruct(false)
			// }
			objFieldPtr = field0
		}
		// addSecurity(servlet, operation) //apix中使用了全局的header，暂时不显示
		var response spec.Response = swagger.getSwaggerResponse(objFieldPtr)
		operation.Responses.StatusCodeResponses[200] = response
		paths[swagger.project.Cfg.SwaggerCfg.UrlPrefix+url] = pathItem
	}
}

//	func addSecurity(function *Function, operation *spec.Operation) {
//		// for _, header := range function.comment.security {
//		// 	operation.Security = append(operation.Security, map[string][]string{
//		// 		header: {"string"},
//		// 	})
//		// }
//		for _, s := range function.pkg.Project.servers {
//			if s.name == function.comment.serverName {
//				for _, filter := range s.urlFilters {
//					url := filter.url
//					filterFunction := filter.function
//					servletUrl := filterFunction.comment.Url
//					if strings.Contains(servletUrl, url) {
//						for _, header := range filterFunction.comment.security {
//							operation.Security = append(operation.Security, map[string][]string{
//								header: {"string"},
//							})
//						}
//					}
//				}
//			}
//		}
//	}
func (swagger *Swagger) GenerateCode(cfg *basic.SwaggerCfg) string {
	project := swagger.project
	for name, pkg := range project.Packages {
		_ = name
		swagger.addServletFromPackage(pkg)
	}
	swaggerJson, err := swagger.swag.MarshalJSON()
	if err != nil {
		fmt.Printf("json.Marshal(s.SwaggerProps) error: %v", err)
		return ""
	}
	if cfg.JsonName != "" {
		//如果不上传，则打印到控制台
		os.WriteFile(cfg.JsonName, swaggerJson, 0644)
	}

	if cfg.Token != "" {
		cmdMap := map[string]any{
			"input": string(swaggerJson),
			"options": map[string]any{
				"targetEndpointFolderId":        cfg.ServletFolder,
				"targetSchemaFolderId":          cfg.SchemaFolder,
				"endpointOverwriteBehavior":     "OVERWRITE_EXISTING",
				"schemaOverwriteBehavior":       "OVERWRITE_EXISTING",
				"updateFolderOfChangedEndpoint": false,
				"prependBasePath":               false,
			},
		}
		data, _ := json.Marshal(cmdMap)
		url := "http://api.apifox.com/v1/projects/" + strconv.Itoa(cfg.ProjectId) + "/import-openapi?locale=zh-CN"
		req, _ := http.NewRequest("POST", url, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Apifox-Api-Version", "2024-03-28")
		req.Header.Set("Authorization", "Bearer "+cfg.Token)
		req.Header.Set("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("error:%v\n", err)
		}
		content, _ := io.ReadAll(response.Body)
		fmt.Printf("response:%v\n", string(content))
		return (string(data))
	}
	return ""
	// fmt.Printf("swagger:%s\n", cmdMap["input"])
}
func (swagger *Swagger) addServletFromPackage(pkg *Package) {
	// swagger.addServletFromFunctionManager(&pkg.FunctionManager)
	for _, class := range pkg.Structs {
		if class.Comment.serverType == Servlet {
			swagger.addServletFromFunctionManager(&class.MethodManager)
		}
	}
}

// 产生schema定义；
// 根据field 逐条产生schema
func (swagger *Swagger) genSchema(class *Struct) map[string]spec.Schema {
	schemas := make(map[string]spec.Schema)
	for _, field := range class.FlatFields() {
		name := field.GetJsonName()
		if name == "-" || name == "" {
			continue
		}

		schema := spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: field.Comment.comment,
			},
		}
		if st, ok := field.Type.(SchemaType); ok {
			st.InitSchema(&schema, swagger)
		} else {
			fmt.Printf("ERROR: field %s::%s %T is not a SchemaType\n", field.Type.IDName(), field.Name, field.Type)
		}
		schemas[name] = schema
	}
	return schemas
}

// 生成struct自己的definition，并保存在最的ref中，方便后续使用；
// ref中记录definition的字符串；  "#/definitions/Node"
//
//	"definitions": {
//		"stuctName" :{}

// 将definition记录在swagger的Definitions中；
func (swagger *Swagger) getStructRef(class *Struct) *spec.Ref {
	if class.ref != nil {
		return class.ref
	}
	ref, _ := spec.NewRef("#/definitions/" + class.StructName)
	class.ref = &ref
	schemas := swagger.genSchema(class)
	result := spec.SchemaProps{
		Type:       []string{"object"},
		Properties: schemas,
	}
	swagger.swag.Definitions[class.StructName] = spec.Schema{
		SchemaProps: result,
	}
	return &ref
}

func (swagger *Swagger) initResponseResult() {
	class := Struct{
		StructName: "ResponseResult",
		Fields: []*Field{
			NewSimpleField(rawTypeMap["int"], "code"),
			NewSimpleField(rawTypeMap["string"], "msg"),
			NewSimpleField(&BaseType{}, "obj"),
		},
	}
	swagger.responseResult = &class
}

func (swagger *Swagger) getSwaggerResponse(objField *Field) spec.Response {
	schema := spec.Schema{
		SchemaProps: spec.SchemaProps{},
	}

	swagger.responseResult.InitSchema(&schema, swagger)
	var result = spec.Response{
		ResponseProps: spec.ResponseProps{
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					AllOf: []spec.Schema{schema},
				},
			},
		},
	}
	if objField == nil {
		return result
	}
	var objSchema = spec.Schema{
		SchemaProps: spec.SchemaProps{},
	}
	basicType := GetBasicType(objField.Type)
	basicType.(SchemaType).InitSchema(&objSchema, swagger)
	ref := spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"object"},
			Properties: map[string]spec.Schema{
				"obj": objSchema,
			},
		},
	}
	result.Schema.AllOf = append(result.Schema.AllOf, ref)
	return result
}
