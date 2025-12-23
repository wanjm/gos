package astinfo

import (
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
		var conflictMap = make(map[string]*NamePair)
		var allColumns []*NamePair

		for _, className := range pkg.SortedStructNames {
			class := pkg.Structs[className]
			if class.Comment.TableName != "" {
				var data = info{
					TableName:    class.StructName,
					RawTableName: class.Comment.TableName,
					DBVariable:   class.Comment.DbVarible,
					Pkg:          &pkg.PkgBasic,
				}
				hasCreateTime := false
				for _, field := range class.Fields {
					if field.Tags["gorm"] == "create_time" {
						hasCreateTime = true
						break
					}
				}
				for _, field := range class.Fields {
					if field.Tags["gorm"] != "" {
						// Collect NamePairs instead of immediately generating columns
						allColumns = append(allColumns, getNamePair(class, "gorm")...)
						if hasCreateTime {
							data.OrderField = "create_time"
							data.OrderDirection = "common.DESCStr"
						} else {
							data.OrderField = "id"
							data.OrderDirection = "common.ASCStr"
						}
						mysqlInfo = append(mysqlInfo, &data)
						break
					} else if field.Tags["bson"] != "" {
						// Collect NamePairs instead of immediately generating columns
						allColumns = append(allColumns, getNamePair(class, "bson")...)
						mongoInfo = append(mongoInfo, &data)
						break
					}
				}
			}
		}

		// Deduplicate all collected columns and generate them
		if len(allColumns) > 0 {
			deduplicatedColumns := DeduplicateNamePairs(conflictMap, allColumns)
			genColumns(file, deduplicatedColumns)
		}

		file.Save()
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
		file.GetImport(astbasic.SimplePackage("go.mongodb.org/mongo-driver/bson", "bson"))
		file.GetImport(astbasic.SimplePackage("go.mongodb.org/mongo-driver/bson/primitive", "primitive"))
		file.GetImport(astbasic.SimplePackage("go.mongodb.org/mongo-driver/mongo", "mongo"))
		file.GetImport(astbasic.SimplePackage("go.mongodb.org/mongo-driver/mongo/options", "options"))
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
func DeduplicateNamePairs(nameMap map[string]*NamePair, namePairs []*NamePair) []*NamePair {
	var result []*NamePair

	for _, pair := range namePairs {
		if existingPair, exists := nameMap[pair.VarName]; exists {
			// Check if the existing pair has the same ColName
			if existingPair.ColName != pair.ColName {
				log.Printf("Warning: NamePair with VarName '%s' already exists with ColName '%s', but new pair has ColName '%s'",
					pair.VarName, existingPair.ColName, pair.ColName)
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
func genColumns(file *astbasic.GenedFile, columns []*NamePair) {
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

// 获取指定tag的值，目前用在gorm和bson两个tag；
// bson:"orgId,omitempty"
// gorm:"column:id;primary_key;AUTO_INCREMENT"
func getNamePair(class *Struct, tag string) []*NamePair {
	var columns []*NamePair
	for _, field := range class.Fields {
		basicType := GetBasicType(field.Type)
		if subClass, ok := basicType.(*Struct); ok {
			if subClass.GoSource.Pkg == class.GoSource.Pkg {
				subColumns := getNamePair(subClass, tag)
				columns = append(columns, subColumns...)
			}
		}
		tag := field.Tags[tag]
		colname := strings.Split(tag, ",")[0]
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
}

func compareInfo(a, b *info) int {
	return strings.Compare(a.TableName, b.TableName)
}

func genMysqlDal(data *info, file *astbasic.GenedFile) {
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

func (a *{{.TableName}}Dal) GetAll(ctx context.Context, options []common.Optioner, cols ...[]string) (item []*{{.Pkg.Name}}.{{.TableName}}, err error) {
	return a.GetLimitAll(ctx, options, 0, cols...)
}
func (a *{{.TableName}}Dal) GetLimitAll(ctx context.Context, options []common.Optioner,count int, cols ...[]string) (item []*{{.Pkg.Name}}.{{.TableName}}, err error) {
	return a.GetLimitAllWithStart(ctx, options, 0, count, cols...)
}
func (a *{{.TableName}}Dal) GetLimitAllWithStart(ctx context.Context, options []common.Optioner,start, count int, cols ...[]string) (item []*{{.Pkg.Name}}.{{.TableName}}, err error) {
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

func (a *{{.TableName}}Dal) List(ctx context.Context, option []common.Optioner, pageNo, pageSize int, cols ...[]string) (list []*{{.Pkg.Name}}.{{.TableName}}, total int64, err error) {
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
			OrderFields: []common.OrderByParam{
				{
					Field:     "{{.OrderField}}",
					Direction: {{.OrderDirection}},
				},
			},
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

func genMongoDal(data *info, file *astbasic.GenedFile) {
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


// Create 创建
func (a *{{.TableName}}Dal) Create(ctx context.Context, item *{{.Pkg.Name}}.{{.TableName}}) error {
	db := a.getDB()
	result, err := db.InsertOne(ctx, item)
	if err != nil {
		common.Error(ctx, "insert record to mongo {{.RawTableName}} failed", common.Err(err))
		return err
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		item.ID = oid
	} else {
		common.Error(ctx, "insert record to mongo {{.TableName}} failed as inserted_id is not objectid")
	}
	return nil
}

func (a *{{.TableName}}Dal) GetAll(ctx context.Context, opts []common.Optioner, cols ...[]string) (item []*{{.Pkg.Name}}.{{.TableName}}, err error) {
	return a.GetLimitAll(ctx, opts, 0, cols...)
}
func (a *{{.TableName}}Dal) GetLimitAll(ctx context.Context, opts []common.Optioner,count int64, cols ...[]string) (item []*{{.Pkg.Name}}.{{.TableName}}, err error) {
	filter := common.GenMongoOption(opts)
	db := a.getDB()
	projection := bson.M{}
	if len(cols) > 0 {
		for _, col := range cols[0] {
			projection[col] = 1
		}
	}
	// 执行查询
	var cur *mongo.Cursor
	cur, err = db.Find(ctx, filter, options.Find().SetProjection(projection).SetLimit(count))

	if err != nil {
		common.Error(ctx, "GetAll from mongo {{.RawTableName}} failed", common.Err(err))
		return nil, err
	}
	defer cur.Close(ctx)
	err = cur.All(ctx, &item)
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

func (a *{{.TableName}}Dal) GetOneById(ctx context.Context, id primitive.ObjectID, cols ...[]string) (item *{{.Pkg.Name}}.{{.TableName}}, err error) {
	return a.GetOne(ctx, []common.Optioner{common.Eq("_id", id)}, cols...)
}

func (a *{{.TableName}}Dal) Set(ctx context.Context, opts []common.Optioner,updates map[string]any) (result *mongo.UpdateResult, err error) {
	return a.Update(ctx, opts, bson.M{"$set": updates})
}

func (a *{{.TableName}}Dal) Update(ctx context.Context, opts []common.Optioner, updates map[string]any) (result *mongo.UpdateResult, err error) {
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
