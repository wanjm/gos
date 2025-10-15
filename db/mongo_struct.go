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
		if len(tableCfg.TableNames) != len(tableCfg.RecordIds) {
			log.Printf("警告: 配置 %v 中的 TableNames 和 RecordIds 数量不匹配，跳过此配置。", tableCfg)
			continue
		}

		pkg := astinfo.GlobalProject.CurrentProject.NewPkgBasic("", tableCfg.ModulePath)
		// 3. 遍历每个表名并调用生成函数
		for i, recordID := range tableCfg.RecordIds {
			if len(recordID) == 0 {
				continue
			}
			tableName := tableCfg.TableNames[i]
			log.Printf("正在为集合 '%s' (ID: %s) 生成结构体...", tableName, recordID)
			// 1. 获取文档
			collection := database.Collection(tableName)
			doc, err := getDocumentByID(collection, recordID)
			if err != nil {
				return fmt.Errorf("获取文档失败: %w", err)
			}
			err = genTableForMongo(tableName, doc, pkg)
			if err != nil {
				log.Printf("为集合 '%s' 生成结构体失败: %v", tableName, err)
				// 选择继续处理下一个表而不是直接返回错误
				continue
			}
			log.Printf("成功为集合 '%s' 生成文件。", tableName)
		}
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
	Name   string
	Fields []FieldInfo
}

const structTpl = `type {{.Name}} struct {
{{- range .Fields}}
	{{.Name}} {{.Type}} ` + "`" + `bson:"{{.BsonTag}},omitempty" json:"{{.JsonTag}}"` + "`" + `
{{- end}}
}`

func genTableForMongo(tableName string, doc bson.M, pkg *astbasic.PkgBasic) error {
	structName := astbasic.ToPascalCase(tableName, true)
	tablepkg := pkg.NewPkgBasic(tableName, "entity/mongo/"+tableName)
	tableFile := tablepkg.NewFile("table")
	return generateStruct(structName, doc, tableFile)
}

func generateStruct(structName string, doc bson.M, tableFile *astbasic.GenedFile) error {
	// 1. Prepare data for the main struct template
	fields := make([]FieldInfo, 0, len(doc))
	keys := make([]string, 0, len(doc))
	for k := range doc {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := doc[key]
		// getGoTypeFromValue will populate nestedStructs as a side effect through recursive calls
		goType := getGoTypeFromValue(key, value, tableFile)
		jsonTag := key
		if key == "_id" {
			jsonTag = "id,omitempty"
		}
		fields = append(fields, FieldInfo{
			Name:    astbasic.ToPascalCase(key, true),
			Type:    goType,
			BsonTag: key,
			JsonTag: jsonTag,
		})
	}

	// 2. Execute the template for the main struct
	t, err := template.New("struct").Parse(structTpl)
	if err != nil {
		return fmt.Errorf("failed to parse struct template: %w", err)
	}

	var mainStructBuilder strings.Builder
	err = t.Execute(&mainStructBuilder, StructInfo{
		Name:   structName,
		Fields: fields,
	})
	if err != nil {
		return fmt.Errorf("failed to execute struct template: %w", err)
	}
	tableFile.AddBuilder(&mainStructBuilder)
	tableFile.Save()
	return nil
}

func getGoTypeFromValue(key string, value interface{}, tableFile *astbasic.GenedFile) string {
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
		structName := astbasic.ToPascalCase(key, true)
		generateStruct(structName, v, tableFile)
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
