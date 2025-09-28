package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/wanjm/gos/basic"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --- 程序主逻辑 ---
// GenTable4DB 是代码生成器的主入口函数
// 它接受 DBConfig 参数，并根据配置循环生成所有表的结构体
func GenTable4DB(config *basic.DBConfig) error {
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

	// 2. 遍历所有表的生成配置
	for _, tableCfg := range config.DbGenCfgs {
		if len(tableCfg.TableNames) != len(tableCfg.RecordIds) {
			log.Printf("警告: 配置 %v 中的 TableNames 和 RecordIds 数量不匹配，跳过此配置。", tableCfg)
			continue
		}

		// 3. 遍历每个表名并调用生成函数
		for i, recordID := range tableCfg.RecordIds {
			if len(recordID) == 0 {
				continue
			}
			tableName := tableCfg.TableNames[i]
			log.Printf("正在为集合 '%s' (ID: %s) 生成结构体...", tableName, recordID)

			err := genTableForMongo(client, config.DBName, tableName, recordID, tableCfg.OutPath)
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

// genTableForMongo 负责为单个 MongoDB 集合生成结构体文件
func genTableForMongo(client *mongo.Client, dbName, collectionName, recordID, outPath string) error {
	// 1. 获取文档
	collection := client.Database(dbName).Collection(collectionName)
	doc, err := getDocumentByID(collection, recordID)
	if err != nil {
		return fmt.Errorf("获取文档失败: %w", err)
	}

	// 2. 准备文件名和路径
	structName := toPascalCase(collectionName)
	snakeName := toSnakeCase(collectionName)
	dirPath := filepath.Join(outPath, snakeName)
	filePath := filepath.Join(dirPath, "table.gen.go") // 文件名固定为 table.go

	// 3. 生成结构体代码字符串
	structCode, err := generateStruct(structName, doc)
	if err != nil {
		return fmt.Errorf("生成结构体代码时出错: %w", err)
	}

	// 4. 组装完整的 Go 文件内容
	var importBuilder strings.Builder
	importBuilder.WriteString("import (\n")
	if strings.Contains(structCode, "time.Time") {
		importBuilder.WriteString("\t\"time\"\n")
	}
	importBuilder.WriteString("\t\"go.mongodb.org/mongo-driver/bson/primitive\"\n")
	importBuilder.WriteString(")\n")
	fileContent := fmt.Sprintf("package %s\n\n%s\n%s", snakeName, importBuilder.String(), structCode)

	// 5. 创建目录并写入文件
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("创建目录 '%s' 失败: %w", dirPath, err)
	}

	if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
		return fmt.Errorf("写入文件 '%s' 失败: %w", filePath, err)
	}

	fmt.Printf("✅ 文件已成功生成: %s\n", filePath)
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

func generateStruct(structName string, doc bson.M) (string, error) {
	nestedStructs := make(map[string]string)

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
		goType := getGoTypeFromValue(key, value, nestedStructs)
		jsonTag := key
		if key == "_id" {
			jsonTag = "id,omitempty"
		}
		fields = append(fields, FieldInfo{
			Name:    toPascalCase(key),
			Type:    goType,
			BsonTag: key,
			JsonTag: jsonTag,
		})
	}

	// 2. Execute the template for the main struct
	t, err := template.New("struct").Parse(structTpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse struct template: %w", err)
	}

	var mainStructBuilder strings.Builder
	err = t.Execute(&mainStructBuilder, StructInfo{
		Name:   structName,
		Fields: fields,
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute struct template: %w", err)
	}

	// 3. Append nested structs
	var finalCodeBuilder strings.Builder
	finalCodeBuilder.WriteString(mainStructBuilder.String())

	if len(nestedStructs) > 0 {
		nestedKeys := make([]string, 0, len(nestedStructs))
		for k := range nestedStructs {
			nestedKeys = append(nestedKeys, k)
		}
		sort.Strings(nestedKeys)
		for _, key := range nestedKeys {
			finalCodeBuilder.WriteString("\n\n")
			finalCodeBuilder.WriteString(nestedStructs[key])
		}
	}

	return finalCodeBuilder.String(), nil
}

func getGoTypeFromValue(key string, value interface{}, nestedStructs map[string]string) string {
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
		structName := toPascalCase(key)
		nestedStructCode, _ := generateStruct(structName, v)
		nestedStructs[structName] = nestedStructCode
		return structName
	case primitive.A:
		if len(v) == 0 {
			return "[]interface{}"
		}
		elemType := getGoTypeFromValue(key, v[0], nestedStructs)
		return "[]" + elemType
	default:
		return "interface{}"
	}
}

func toPascalCase(s string) string {
	if s == "_id" {
		return "ID"
	}
	var result strings.Builder
	capitalizeNext := true
	for _, r := range s {
		if r == '_' || r == '-' {
			capitalizeNext = true
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if capitalizeNext {
				result.WriteRune(unicode.ToUpper(r))
				capitalizeNext = false
			} else {
				result.WriteRune(r)
			}
		}
	}
	return result.String()
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
