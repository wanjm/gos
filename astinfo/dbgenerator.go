package astinfo

import (
	"fmt"
	"html/template"
	"log"
	"path/filepath"
	"slices"
	"strings"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/basic"
)

type DbManager struct {
}

func TruncateAtLastEntity(s string) string {
	// 找到最后一个"entity"的起始索引
	lastIndex := strings.LastIndex(s, "entity")
	if lastIndex == -1 {
		return s
	}
	// 从最后一个"entity"的起始位置截断，保留到该位置（不包含"entity"本身）
	// 如果需要包含"entity"，则截断索引应为 lastIndex + len("entityentity")
	return s[:lastIndex]
}

// 此处的目录结构可能是
// entity/mysql/tableName/table.go;
// 但是无论是哪种格式，我们都需要在entity父目录下创建同级的dal目录；
func (db *DbManager) Gen() {
	var pkgMysqlMap = map[string][]*info{}
	var pkgMongoMap = map[string][]*info{}
	//key
	pkgs := GlobalProject.Packages
	for _, pkg := range pkgs {
		var mysqlInfo []*info
		var mongoInfo []*info
		file := pkg.NewFile("column")
		listFile := pkg.NewFile("list")
		hasList := false
		var conflictMap = make(map[string]*NamePair)
		var allColumns []*NamePair

		for _, className := range pkg.SortedStructNames {
			class := pkg.Structs[className]
			if class.Comment.TableName != "" {
				genEntityList(class, listFile)
				hasList = true
				var data = info{
					TableName:    class.StructName,
					RawTableName: class.Comment.TableName,
					DBVariable:   class.Comment.DbVarible,
					Pkg:          &pkg.PkgBasic,
				}
				hasCreateTime := false
				hasId := false
				for _, field := range class.Fields {
					// 此处已经进行了column的处理
					colName := field.DbColumnName
					if colName == CreateTime {
						hasCreateTime = true
					}
					if colName == Id {
						hasId = true
					}
				}
				for _, field := range class.Fields {
					if field.Tags[GORM] != "" {
						// Collect NamePairs instead of immediately generating columns
						allColumns = append(allColumns, getNamePair(class)...)
						if hasCreateTime {
							data.OrderField = CreateTime
							data.OrderDirection = "common.DESCStr"
						} else if hasId {
							data.OrderField = Id
							data.OrderDirection = "common.ASCStr"
						}
						// else keep OrderField empty
						mysqlInfo = append(mysqlInfo, &data)
						break
					} else if field.Tags[BSON] != "" {
						// Collect NamePairs instead of immediately generating columns
						allColumns = append(allColumns, getNamePair(class)...)
						data.IDName = getIdName(class)
						mongoInfo = append(mongoInfo, &data)
						break
					}
				}
			}
		}

		// Deduplicate all collected columns and generate them
		if len(allColumns) > 0 {
			deduplicatedColumns := DeduplicateNamePairs(conflictMap, allColumns, pkg.FilePath)
			genColumns(file, deduplicatedColumns)
		}

		file.Save()
		if hasList {
			listFile.Save()
		}
		if len(mysqlInfo) == 0 && len(mongoInfo) == 0 {
			continue
		}

		dalPath := TruncateAtLastEntity(pkg.FilePath)
		if len(mysqlInfo) > 0 {
			pkgMysqlMap[dalPath] = append(pkgMysqlMap[dalPath], mysqlInfo...)
		}
		if len(mongoInfo) > 0 {
			pkgMongoMap[dalPath] = append(pkgMongoMap[dalPath], mongoInfo...)
		}
	}
	for pkgPath, mysqlInfo := range pkgMysqlMap {
		//由于保存文件，仅仅使用FilePath；
		pkg := astbasic.PkgBasic{
			Name:     "dal",
			FilePath: filepath.Join(pkgPath, "dal"),
		}
		file := pkg.NewFile("mysql.dal")
		file.GetImport(astbasic.SimplePackage(basic.Cfg.Generation.CommonMod, "common"))
		file.GetImport(astbasic.SimplePackage("context", "context"))
		file.GetImport(astbasic.SimplePackage("gorm.io/gorm", "gorm"))
		if len(mysqlInfo) > 1 {
			slices.SortFunc(mysqlInfo, compareInfo)
		}
		for _, info := range mysqlInfo {
			genMysqlDal(info, file)
		}
		file.Save()
	}
	for pkgPath, mongoInfo := range pkgMongoMap {
		pkg := astbasic.PkgBasic{
			Name:     "dal",
			FilePath: filepath.Join(pkgPath, "dal"),
		}
		file := pkg.NewFile("mongo.dal")
		file.GetImport(astbasic.SimplePackage("context", "context"))
		file.GetImport(astbasic.SimplePackage(basic.Cfg.Generation.CommonMod, "common"))
		// file.GetImport(astbasic.SimplePackage("go.mongodb.org/mongo-driver/bson", "bson"))
		file.GetImport(astbasic.SimplePackage("go.mongodb.org/mongo-driver/bson/primitive", "primitive"))
		file.GetImport(astbasic.SimplePackage("go.mongodb.org/mongo-driver/mongo", "mongo"))
		// file.GetImport(astbasic.SimplePackage("go.mongodb.org/mongo-driver/mongo/options", "options"))
		if len(mongoInfo) > 1 {
			slices.SortFunc(mongoInfo, compareInfo)
		}
		for _, info := range mongoInfo {
			genMongoDal(info, file)
		}
		file.Save()
	}
}

type tbInfo struct {
}
type NamePair struct {
	VarName string
	ColName string
}

// DeduplicateNamePairs takes a map and a slice of NamePair pointers,
// returns a deduplicated slice of NamePair pointers.
// Uses VarName as the key for the map to check for duplicates.
// If a NamePair with the same VarName exists but has different ColName, prints a warning.
func DeduplicateNamePairs(nameMap map[string]*NamePair, namePairs []*NamePair, filePath string) []*NamePair {
	var result []*NamePair

	for _, pair := range namePairs {
		if existingPair, exists := nameMap[pair.VarName]; exists {
			// Check if the existing pair has the same ColName
			if existingPair.ColName != pair.ColName {
				log.Printf("Warning: NamePair with VarName '%s' already exists with ColName '%s', but new pair has ColName '%s' in package %s",
					pair.VarName, existingPair.ColName, pair.ColName, filePath)
			}
			// Skip this pair as it already exists
			continue
		}

		// Add to map and result
		nameMap[pair.VarName] = pair
		result = append(result, pair)
	}

	return result
}

// 这样写，是为了解决一个entity目录中有多个表的情况；
// 但是还需要解决多个表列名重复的问题， 所以最终还是需要一个entity一个目录；
// 为了兼容多个表的情况，代码需要改为先产生namePair，然后在产生column的情况，这样可以去重；但名字相同，列名不同时，可以报错；
func genColumns(file *astbasic.GenedGoFile, columns []*NamePair) {
	tmplText := `
	const (
	{{range .}}
	{{.VarName}} = "{{.ColName}}"
	{{end}}
	)
	`
	tmpl, err := template.New("personInfo").Parse(tmplText)
	if err != nil {
		log.Fatalf("解析模板失败: %v", err)
	}
	var sb strings.Builder
	err = tmpl.Execute(&sb, columns)
	if err != nil {
		log.Fatalf("执行模板失败: %v", err)
	}
	file.AddBuilder(&sb)
}

// getIdName 获取id的变量名，如果id不是系统默认类型，则返回空字符串；
func getIdName(class *Struct) string {
	for _, field := range class.Fields {
		if field.DbColumnName == "_id" {
			// 简单检查是否是ObjectID类型；
			if field.Type.RefName(nil) == "ObjectID" {
				return field.Name
			}
			return ""
		}
	}
	return ""
}

// 获取指定tag的值，目前用在gorm和bson两个tag；
// bson:"orgId,omitempty"
// gorm:"column:id;primary_key;AUTO_INCREMENT"
func getNamePair(class *Struct) []*NamePair {
	var columns []*NamePair
	for _, field := range class.Fields {
		basicType := GetRootBasicType(field.Type)
		if subClass, ok := basicType.(*Struct); ok {
			if subClass.GoSource.Pkg == class.GoSource.Pkg {
				subColumns := getNamePair(subClass)
				columns = append(columns, subColumns...)
			}
		}
		colname := field.DbColumnName
		if colname != "" && colname != "-" {
			columns = append(columns, &NamePair{
				VarName: "C_" + field.Name,
				ColName: colname,
			})
		}
	}
	return columns
}

type info struct {
	TableName      string
	RawTableName   string
	DBVariable     string
	Pkg            *astbasic.PkgBasic
	OrderField     string
	OrderDirection string
	IDName         string // mongo中如果_id不是系统默认类型，则插入语句要少生成代码, 否则使用正确的类型；
}

func compareInfo(a, b *info) int {
	return strings.Compare(a.TableName, b.TableName)
}

func genMysqlDal(data *info, file *astbasic.GenedGoFile) {
	codeTemplate := `
// {{.RawTableName}}
//
// @gos autogen
type {{.TableName}}Dal struct {
	{{.DBVariable}} *gorm.DB
}

func (a *{{.TableName}}Dal) getDB(ctx context.Context) *gorm.DB {
	return a.{{.DBVariable}}.WithContext(ctx).Table("{{.RawTableName}}")
}
	
func (c *{{.TableName}}Dal) getDBOperation(context context.Context) common.DbOperation {
	return common.DbOperation{
		Db:        c.{{.DBVariable}},
		TableName: "{{.RawTableName}}",
		Context:   context,
	}
}

// Create 创建
func (a *{{.TableName}}Dal) Create(ctx context.Context, item *{{.Pkg.Name}}.{{.TableName}}) error {
	dbOperation := a.getDBOperation(ctx)
	err := dbOperation.Create(item)
	if err != nil {
		common.Error(ctx, "insert data to {{.RawTableName}} failed", common.Err(err))
	}
	return err
}

func (a *{{.TableName}}Dal) GetAll(ctx context.Context, options []common.Optioner, cols ...[]string) (item {{.Pkg.Name}}.{{.TableName}}List, err error) {
	return a.GetLimitAll(ctx, options, 0, cols...)
}
func (a *{{.TableName}}Dal) GetLimitAll(ctx context.Context, options []common.Optioner,count int, cols ...[]string) (item {{.Pkg.Name}}.{{.TableName}}List, err error) {
	return a.GetLimitAllWithStart(ctx, options, 0, count, cols...)
}
func (a *{{.TableName}}Dal) GetLimitAllWithStart(ctx context.Context, options []common.Optioner,start, count int, cols ...[]string) (item {{.Pkg.Name}}.{{.TableName}}List, err error) {
	var colNames []string
	if len(cols) > 0 {
		colNames = cols[0]
	}
	dbOperation := a.getDBOperation(ctx)
	err = dbOperation.Query(
		&common.SqlQueryOptions{
			QueryFields: options,
			Offset:      start,
			Limit:       count,
			SelectFields: colNames,
		},
		&item,
	)
	if err != nil {
		//添加log，打印错误日志；
		common.Error(ctx, "GetAll DB record from {{.RawTableName}} failed", common.Err(err))
	}
	return
}

func (a *{{.TableName}}Dal) GetOne(ctx context.Context, options []common.Optioner, cols ...[]string) (item *{{.Pkg.Name}}.{{.TableName}}, err error) {
	res, err := a.GetLimitAll(ctx, options, 1, cols...)
	if err != nil {
		return
	}
	if len(res) > 0 {
		item = res[0]
	}
	return
}

func (a *{{.TableName}}Dal) GetOneById(ctx context.Context, id int32, cols ...[]string) (item *{{.Pkg.Name}}.{{.TableName}}, err error) {
	return a.GetOne(ctx, []common.Optioner{common.Eq("id", id)}, cols...)
}

func (a *{{.TableName}}Dal) List(ctx context.Context, option []common.Optioner, pageNo, pageSize int, cols ...[]string) (list {{.Pkg.Name}}.{{.TableName}}List, total int64, err error) {
	var colNames []string
	if len(cols) > 0 {
		colNames = cols[0]
	}	
    dbop := a.getDBOperation(ctx)
	err = dbop.QueryCV(
		&common.SqlQueryOptions{
			QueryFields: option,
			Offset:      int(pageNo * pageSize),
			Limit:       int(pageSize),
			{{if .OrderField}}
			OrderFields: []common.OrderByParam{
				{
					Field:     "{{.OrderField}}",
					Direction: {{.OrderDirection}},
				},
			},
			{{end}}
			SelectFields: colNames,
		},
		&total,
		&list,
	)
	if err != nil {
		//添加log，打印错误日志；
		common.Error(ctx, "List record of {{.RawTableName}} failed", common.Err(err))
	}
	return
}

// Update
func (a *{{.TableName}}Dal) Update(ctx context.Context, options []common.Optioner,updates map[string]any) (err error) {
	op := a.getDBOperation(ctx)
	err = op.Update(&common.SqlUpdateOptions{
		QueryFields: options,
		Updates:     updates,
	})
	if err != nil {
		//添加log，打印错误日志；
		common.Error(ctx, "Update record of {{.RawTableName}} failed", common.Err(err))
	}
	return
}

func (a *{{.TableName}}Dal) UpdateById(ctx context.Context, id int32, updates map[string]any) error {
	return a.Update(ctx, []common.Optioner{common.Eq("id", id)}, updates)
}

func (a *{{.TableName}}Dal) Delete(ctx context.Context, options []common.Optioner) error {
	op := a.getDBOperation(ctx)
	err := op.Delete(options)
	if err != nil {
		common.Error(ctx, "Delete record of {{.RawTableName}} failed", common.Err(err))
	}
	return err
}

func (a *{{.TableName}}Dal) DeleteByIds(ctx context.Context, ids []int32) error {
	return a.Delete(ctx, []common.Optioner{common.In("id", ids)})
}
`
	var content strings.Builder
	tpl, err := template.New("common").Parse(codeTemplate)
	if err != nil {
		// 处理模板解析错误
		panic(err)
	}
	if err := tpl.Execute(&content, data); err != nil {
		// 处理模板执行错误
		panic(err)
	}
	file.GetImport(data.Pkg)
	file.AddBuilder(&content)
}

func genMongoDal(data *info, file *astbasic.GenedGoFile) {
	codeTemplate := `
// {{.RawTableName}}
//
// @gos autogen
type {{.TableName}}Dal struct {
	{{.DBVariable}} *mongo.Database
	DbName string "default:\"{{.RawTableName}}\""
}

func (a *{{.TableName}}Dal) getDB() *mongo.Collection {
	return a.{{.DBVariable}}.Collection(a.DbName)
}

func (a *{{.TableName}}Dal) getOperation(ctx context.Context, opts []common.Optioner) *common.MongoQueryOperation {
	db := a.getDB()
	return common.NewMongoQueryOperation(ctx, db, opts)
}

// Create 创建
func (a *{{.TableName}}Dal) Create(ctx context.Context, item *{{.Pkg.Name}}.{{.TableName}}) error {
	db := a.getDB()
	result, err := db.InsertOne(ctx, item)
	if err != nil {
		common.Error(ctx, "insert record to mongo {{.RawTableName}} failed", common.Err(err))
		return err
	}
	{{if not (eq .IDName "")}}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		item.{{.IDName}} = oid
	} else {
		common.Error(ctx, "insert record to mongo {{.TableName}} failed as inserted_id is not objectid")
	}
	{{else}}
		_ = result
	{{end}}
	return nil
}

func (a *{{.TableName}}Dal) GetAll(ctx context.Context, opts []common.Optioner, cols ...[]string) (item {{.Pkg.Name}}.{{.TableName}}List, err error) {
	return a.GetLimitAll(ctx, opts, 0, cols...)
}
func (a *{{.TableName}}Dal) GetLimitAll(ctx context.Context, opts []common.Optioner,count int64, cols ...[]string) (item {{.Pkg.Name}}.{{.TableName}}List, err error) {
	op := a.getOperation(ctx, opts)
	op.SetLimit(count)
	if len(cols) > 0 {
		op.SetProjection(cols[0])
	}
	err = op.Query(&item)
	if err != nil {
		common.Error(ctx, "GetAll from mongo {{.RawTableName}} failed when call all/decode", common.Err(err))
		return nil, err
	}
	return
}

func (a *{{.TableName}}Dal) List(ctx context.Context, opts []common.Optioner, pageNum int, pageSize int, cols ...[]string) (list {{.Pkg.Name}}.{{.TableName}}List, total int64, err error) {
	op := a.getOperation(ctx, opts)
	total, err = op.Count()
	if err != nil {
		common.Error(ctx, "count {{.RawTableName}} failed", common.Err(err))
		return nil, 0, err
	}
	// Get paginated results
	skip := int64(pageNum * pageSize)
	limit := int64(pageSize)
	op.SetSkip(skip).SetLimit(limit)
	if len(cols) > 0 {
		op.SetProjection(cols[0])
	}
	err = op.Query(&list)
	if err != nil {
		common.Error(ctx, "query {{.RawTableName}} failed", common.Err(err))
		return nil, 0, err
	}
	return list, total, nil
}

func (a *{{.TableName}}Dal) GetOne(ctx context.Context, options []common.Optioner, cols ...[]string) (item *{{.Pkg.Name}}.{{.TableName}}, err error) {
	res, err := a.GetLimitAll(ctx, options, 1, cols...)
	if err != nil {
		return
	}
	if len(res) > 0 {
		item = res[0]
	}
	return
}

func (a *{{.TableName}}Dal) GetOneById(ctx context.Context, id primitive.ObjectID, cols ...[]string) (item *{{.Pkg.Name}}.{{.TableName}}, err error) {
	return a.GetOne(ctx, []common.Optioner{common.Eq("_id", id)}, cols...)
}

func (a *{{.TableName}}Dal) Set(ctx context.Context, opts []common.Optioner,updates map[string]any) (result *mongo.UpdateResult, err error) {
	return a.Update(ctx, opts, common.MongoMap{"$set": updates})
}
// update 支持bit位等操作。如果仅仅简单更新，且需要更简单，需要Set方法；
func (a *{{.TableName}}Dal) Update(ctx context.Context, opts []common.Optioner, updates common.MongoMap) (result *mongo.UpdateResult, err error) {
	filter := common.GenMongoOption(opts)
	db := a.getDB()
	result, err = db.UpdateMany(ctx, filter, updates)
	if err != nil {
		common.Error(ctx, "update mongo record {{.RawTableName}} failed", common.Err(err))
	}
	return
}
`
	tpl, err := template.New("common").Parse(codeTemplate)
	if err != nil {
		// 处理模板解析错误
		panic(err)
	}
	var content strings.Builder
	if err := tpl.Execute(&content, data); err != nil {
		// 处理模板执行错误
		panic(err)
	}
	file.GetImport(data.Pkg)
	file.AddBuilder(&content)
}

func genEntityList(class *Struct, file *astbasic.GenedGoFile) {
	// 1. Generate List Type
	// type {Entity}List []*Entity
	entityName := class.StructName
	listType := entityName + "List"

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\ntype %s []*%s\n", listType, entityName))

	// Helper to find field
	findField := func(dbColumnName string) *Field {
		for _, f := range class.Fields {
			if f.DbColumnName == dbColumnName {
				return f
			}
		}
		return nil
	}

	// 2. Generate Arrays methods
	for _, columnName := range class.Comment.Arrays {
		field := findField(columnName)
		if field == nil {
			log.Printf("Warning: Field %s not found in struct %s for array generation", columnName, entityName)
			continue
		}

		fieldType := field.Type.RefName(file)
		// Get{FieldName}List
		methodName := "Get" + field.Name + "List"

		sb.WriteString(fmt.Sprintf("\nfunc (l %s) %s() []%s {\n", listType, methodName, fieldType))
		sb.WriteString(fmt.Sprintf("\tdata := make([]%s, 0, len(l))\n", fieldType))
		sb.WriteString("\tfor _, item := range l {\n")
		sb.WriteString("\t\tif item != nil {\n")
		sb.WriteString(fmt.Sprintf("\t\t\tdata = append(data, item.%s)\n", field.Name))
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t}\n")
		sb.WriteString("\treturn data\n")
		sb.WriteString("}\n")
	}

	// 3. Generate Maps methods
	for _, columnName := range class.Comment.Maps {
		field := findField(columnName)
		if field == nil {
			log.Printf("Warning: Field %s not found in struct %s for map generation", columnName, entityName)
			continue
		}

		fieldType := field.Type.RefName(file)
		// GetMapBy{FieldName}
		methodName := "GetMapBy" + field.Name

		sb.WriteString(fmt.Sprintf("\nfunc (l %s) %s() map[%s]*%s {\n", listType, methodName, fieldType, entityName))
		sb.WriteString(fmt.Sprintf("\tdata := make(map[%s]*%s, len(l))\n", fieldType, entityName))
		sb.WriteString("\tfor _, item := range l {\n")
		sb.WriteString("\t\tif item != nil {\n")
		sb.WriteString(fmt.Sprintf("\t\t\tdata[item.%s] = item\n", field.Name))
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t}\n")
		sb.WriteString("\treturn data\n")
		sb.WriteString("}\n")
	}

	file.AddBuilder(&sb)
}
