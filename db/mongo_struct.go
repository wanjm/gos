package db

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/astinfo"
	"github.com/wanjm/gos/basic"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --- 程序主逻辑 ---
// GenTable4Mongo 是代码生成器的主入口函数
// 它接受 DBConfig 参数，并根据配置循环生成所有表的结构体
func GenTableFromMongo(config *basic.DBConfig, moduleMap map[string]struct{}) error {
	if strings.ToLower(config.DBType) != "mongo" {
		return fmt.Errorf("当前仅支持 'mongo' 数据库类型, 您提供的是 '%s'", config.DBType)
	}

	// 1. 建立数据库连接
	client, err := connectToMongo(config.DSN)
	if err != nil {
		return fmt.Errorf("无法连接到 MongoDB: %w", err)
	}
	defer client.Disconnect(context.Background())
	log.Println("成功连接到 MongoDB!")
	database := client.Database(config.DBName)
	// 2. 遍历所有表的生成配置
	for _, tableCfg := range config.DbGenCfgs {
		pkg := astinfo.GlobalProject.CurrentProject.NewPkgBasic("", tableCfg.ModulePath)
		entityPkg := pkg.NewPkgBasic("entity", "entity")
		file := entityPkg.NewFile("mongo.alias")
		var aliasStringBuilder strings.Builder
		// 3. 遍历每个表并调用生成函数
		for _, t := range tableCfg.Tables {
			tableName := t.Name
			recordIDs := t.RecordIds
			if len(recordIDs) == 0 {
				log.Printf("警告: 表 '%s' 没有 RecordIds，跳过。", tableName)
				continue
			}
			recordID := recordIDs[0]
			log.Printf("正在为集合 '%s' (ID: %s) 生成结构体...", tableName, recordID)
			collection := database.Collection(tableName)
			doc, err := getDocumentByID(collection, recordID)
			if err != nil {
				return fmt.Errorf("获取文档失败: %w", err)
			}
			err = genTableForMongo(doc, pkg, config.DBName, t)
			if err != nil {
				log.Printf("为集合 '%s' 生成结构体失败: %v", tableName, err)
				continue
			}
			structName := astbasic.ToCamelCase(tableName, true)
			tablePkg := pkg.NewPkgBasic(tableName, "entity/mongo/"+tableName)
			file.GetImport(tablePkg)
			aliasStringBuilder.WriteString("type " + structName + " = " + tableName + "." + structName + "\n")
			log.Printf("成功为集合 '%s' 生成文件。", tableName)
		}
		file.AddBuilder(&aliasStringBuilder)
		file.Save()
	}
	return nil
}

// --- 辅助函数 (基本与上一版相同，增加了 toSnakeCase) ---

func connectToMongo(dsn string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		return nil, err
	}
	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("无法 Ping通 MongoDB: %w", err)
	}
	return client, nil
}

func getDocumentByID(collection *mongo.Collection, id string) (bson.M, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("无效的 ObjectID 格式: %w", err)
	}
	var result bson.M
	filter := bson.M{"_id": objectID}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = collection.FindOne(ctx, filter).Decode(&result)
	return result, err
}

// FieldInfo holds information about a single struct field for template generation.
type FieldInfo struct {
	Name    string
	Type    string
	BsonTag string
	JsonTag string
}

// StructInfo holds information about a struct for template generation.
type StructInfo struct {
	Name       string
	DBVariable string
	Fields     []FieldInfo
}

const structTpl = `type {{.Name}} struct {
{{- range .Fields}}
	{{.Name}} {{.Type}} ` + "`" + `bson:"{{.BsonTag}},omitempty" json:"{{.JsonTag}}"` + "`" + `
{{- end}}
}
`

func genTableForMongo(doc bson.M, pkg *astbasic.PkgBasic, dbVariable string, t basic.TableCfg) error {
	tableName := t.Name
	structName := astbasic.ToCamelCase(tableName, true)
	tablepkg := pkg.NewPkgBasic(tableName, "entity/mongo/"+tableName)
	tableFile := tablepkg.NewFile("table")
	var gosStringBuilder strings.Builder
	gosStringBuilder.WriteString("// @gos tblName=")
	gosStringBuilder.WriteString(tableName)
	gosStringBuilder.WriteString(" dbVariable=")
	gosStringBuilder.WriteString(dbVariable)
	if len(t.Arrays) > 0 {
		gosStringBuilder.WriteString(" arrays=")
		gosStringBuilder.WriteString(strings.Join(t.Arrays, ","))
	}
	if len(t.Maps) > 0 {
		gosStringBuilder.WriteString(" maps=")
		gosStringBuilder.WriteString(strings.Join(t.Maps, ","))
	}
	gosStringBuilder.WriteString("\n")
	tableFile.AddBuilder(&gosStringBuilder)
	fields, err := generateStruct(structName, doc, tableFile)
	if err != nil {
		return err
	}
	var methods []string
	for _, a := range t.Arrays {
		methods = append(methods, a+"s")
	}
	for _, m := range t.Maps {
		methods = append(methods, m+"Map")
	}

	return genTableGenForMongo(tablepkg, structName, methods, fields)
}

func generateStruct(structName string, doc bson.M, tableFile *astbasic.GenedGoFile) ([]FieldInfo, error) {
	// 1. Prepare data for the main struct template
	fields := make([]FieldInfo, 0, len(doc))
	keys := make([]string, 0, len(doc))
	for k := range doc {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := doc[key]
		goType := getGoTypeFromValue(key, value, tableFile)
		jsonTag := key
		if key == "_id" {
			jsonTag = "id,omitempty"
		}
		fields = append(fields, FieldInfo{
			Name:    astbasic.ToCamelCase(key, true),
			Type:    goType,
			BsonTag: key,
			JsonTag: jsonTag,
		})
	}

	// 2. Execute the template for the main struct
	t, err := template.New("struct").Parse(structTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse struct template: %w", err)
	}

	var mainStructBuilder strings.Builder
	err = t.Execute(&mainStructBuilder, StructInfo{
		Name:   structName,
		Fields: fields,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute struct template: %w", err)
	}
	tableFile.AddBuilder(&mainStructBuilder)
	tableFile.Save()
	return fields, nil
}

// genTableGenForMongo generates table.gen.go with collection type and Xs/Xmap methods.
func genTableGenForMongo(tablepkg *astbasic.PkgBasic, structName string, methods []string, fields []FieldInfo) error {
	fieldMap := make(map[string]string)
	for _, f := range fields {
		fieldMap[f.Name] = f.Type
	}
	genFile := tablepkg.NewFile("table.gen")
	var sb strings.Builder
	sb.WriteString("// Code generated by gos DO NOT EDIT.\n\n")
	sb.WriteString("// Collection type\n")
	collectionType := structName + "s"
	sb.WriteString("type " + collectionType + " []*" + structName + "\n\n")
	for _, methodName := range methods {
		fieldName, isMap := parseMethodName(methodName)
		if fieldName == "" {
			continue
		}
		fieldType, ok := fieldMap[fieldName]
		if !ok {
			continue
		}
		if isMap {
			sb.WriteString("// Xmap: returns map from field value to *" + structName + "\n")
			sb.WriteString("func (u " + collectionType + ") " + methodName + "() map[" + fieldType + "]*" + structName + " {\n")
			sb.WriteString("\tout := make(map[" + fieldType + "]*" + structName + ", len(u))\n")
			sb.WriteString("\tfor _, e := range u {\n")
			sb.WriteString("\t\tif e != nil {\n")
			sb.WriteString("\t\t\tout[e." + fieldName + "] = e\n")
			sb.WriteString("\t\t}\n")
			sb.WriteString("\t}\n")
			sb.WriteString("\treturn out\n")
		} else {
			sb.WriteString("// Xs: returns slice of field values (same type as field)\n")
			sb.WriteString("func (u " + collectionType + ") " + methodName + "() []" + fieldType + " {\n")
			sb.WriteString("\tout := make([]" + fieldType + ", 0, len(u))\n")
			sb.WriteString("\tfor _, e := range u {\n")
			sb.WriteString("\t\tif e != nil {\n")
			sb.WriteString("\t\t\tout = append(out, e." + fieldName + ")\n")
			sb.WriteString("\t\t}\n")
			sb.WriteString("\t}\n")
			sb.WriteString("\treturn out\n")
		}
		sb.WriteString("}\n\n")
	}
	genFile.AddBuilder(&sb)
	genFile.Save()
	return nil
}

func getGoTypeFromValue(key string, value interface{}, tableFile *astbasic.GenedGoFile) string {
	if value == nil {
		return "interface{}"
	}
	switch v := value.(type) {
	case primitive.ObjectID:
		return "primitive.ObjectID"
	case primitive.DateTime:
		return "time.Time"
	case string:
		return "string"
	case bool:
		return "bool"
	case int32:
		return "int32"
	case int64:
		return "int64"
	case float64:
		return "float64"
	case bson.M:
		structName := astbasic.ToCamelCase(key, true)
		_, _ = generateStruct(structName, v, tableFile)
		return structName
	case primitive.A:
		if len(v) == 0 {
			return "[]interface{}"
		}
		elemType := getGoTypeFromValue(key, v[0], tableFile)
		return "[]" + elemType
	default:
		return "interface{}"
	}
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
