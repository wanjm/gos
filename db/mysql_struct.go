package db

import (
	"database/sql"
	"fmt"
	"strings"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/astinfo"
	"github.com/wanjm/gos/basic"
)

func GenTableFromMySQL(config *basic.DBConfig, moduleMap map[string]struct{}) error {
	// 1. Connect to MySQL
	db, err := connectToMySQL(config.DSN)
	if err != nil {
		return fmt.Errorf("无法连接到 MySQL: %w", err)
	}
	defer db.Close()
	for _, cfg := range config.DbGenCfgs {
		cfg.DBName = config.DBName
		if _, ok := moduleMap[cfg.ModulePath]; ok {
			genTable(cfg, db)
		}
	}
	return nil
}

// GenTableFromMySQL connects to MySQL, gets the DDL of a table, and generates a Go struct definition.
func genTable(tableCfg *basic.TableGenCfg, db *sql.DB) error {
	pkg := astinfo.GlobalProject.CurrentProject.NewPkgBasic("", tableCfg.ModulePath)
	entityPkg := pkg.NewPkgBasic("entity", "entity")
	file := entityPkg.NewFile("mysql.alias")
	var aliasStringBuilder strings.Builder
	for _, tableName := range tableCfg.TableNames {
		// 2. Get DDL
		ddl, err := getTableDDL(db, tableName)
		if err != nil {
			fmt.Printf("获取表 DDL 失败: %v\n", err)
			return err
		}
		// 3. Parse DDL and generate struct
		tablepkg := entityPkg.NewPkgBasic(tableName, "mysql/"+tableName)
		file.GetImport(tablepkg)
		structName, err := GenerateStructFromDDL(tableName, ddl, tablepkg, tableCfg.DBName)
		if err != nil {
			fmt.Printf("生成结构体代码失败: %v\n", err)
			return err
		}
		aliasStringBuilder.WriteString("type " + structName + "= " + tableName + "." + structName + "\n")
	}
	file.AddBuilder(&aliasStringBuilder)
	file.Save()
	return nil
}

// connectToMySQL creates a connection to MySQL using DSN
func connectToMySQL(dsn string) (*sql.DB, error) {
	return sql.Open("mysql", dsn)
}

// getTableDDL retrieves the CREATE TABLE statement for a given table
func getTableDDL(db *sql.DB, tableName string) (string, error) {
	var table string
	var ddl string
	row := db.QueryRow("SHOW CREATE TABLE `" + tableName + "`")
	err := row.Scan(&table, &ddl)
	return ddl, err
}

// GenerateStructFromDDL parses the DDL and generates a Go struct definition
// return Struct Name;
func GenerateStructFromDDL(tableName, ddl string, tablepkg *astbasic.PkgBasic, dbVariable string) (string, error) {
	tableFile := tablepkg.NewFile("table")
	// Simple parser: extract column lines from DDL
	lines := strings.Split(ddl, "\n")
	type fieldInfo struct {
		Name    string
		Type    string
		JsonTag string
		GormTag string
		Comment string
	}
	var fields []fieldInfo
	var tableComment string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "`") {
			parts := strings.SplitN(line, " ", 3)
			if len(parts) < 2 {
				continue
			}
			colName := strings.Trim(parts[0], "`")
			colType := parts[1]
			var comment string
			if commentIndex := strings.Index(line, "COMMENT '"); commentIndex != -1 {
				commentPart := line[commentIndex+len("COMMENT '"):]
				if endQuoteIndex := strings.Index(commentPart, "'"); endQuoteIndex != -1 {
					comment = commentPart[:endQuoteIndex]
				}
			}
			goType := mysqlTypeToGoType(colType, tableFile)
			fieldName := astbasic.ToCamelCase(colName, true)
			jsonTag := astbasic.ToCamelCase(colName, false)
			gormTag := colName
			fields = append(fields, fieldInfo{
				Name:    fieldName,
				Type:    goType,
				JsonTag: jsonTag,
				GormTag: gormTag,
				Comment: comment,
			})
		} else if strings.HasPrefix(line, ")") {
			if commentIndex := strings.Index(line, "COMMENT='"); commentIndex != -1 {
				commentPart := line[commentIndex+len("COMMENT='"):]
				if endQuoteIndex := strings.Index(commentPart, "'"); endQuoteIndex != -1 {
					tableComment = commentPart[:endQuoteIndex]
				}
			}
		}
	}
	structName := astbasic.ToCamelCase(tableName, true)
	var structComment string
	if tableComment != "" {
		structComment = fmt.Sprintf("// %s %s\n", structName, tableComment)
	}
	const structTpl = `
	{{.StructComment}}
	// @gos tblName={{.TableName}} dbVariable={{.DbVariable}}
type {{.StructName}} struct {
{{range .Fields}}
	{{.Name}} {{.Type}} "json:\"{{.JsonTag}}\" gorm:\"column:{{.GormTag}}\"" // {{.Comment}}
{{end}}
}
`

	tpl, err := template.New("struct").Parse(structTpl)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	err = tpl.Execute(&sb, map[string]interface{}{
		"StructComment": structComment,
		"StructName":    structName,
		"Fields":        fields,
		"TableName":     tableName,
		"DbVariable":    dbVariable,
	})
	if err != nil {
		return "", err
	}
	tableFile.AddBuilder(&sb)
	tableFile.Save()
	return structName, nil
}

// mysqlTypeToGoType maps MySQL types to Go types (basic mapping)
func mysqlTypeToGoType(mysqlType string, file *astbasic.GenedFile) string {
	t := strings.ToLower(mysqlType)
	switch {
	case strings.HasPrefix(t, "int"):
		return "int32"
	case strings.HasPrefix(t, "bigint"):
		return "int64"
	// case strings.HasPrefix(t, "tinyint(1)"):
	// return "bool"
	case strings.HasPrefix(t, "tinyint"):
		return "int8"
	case strings.HasPrefix(t, "smallint"):
		return "int16"
	case strings.HasPrefix(t, "mediumint"):
		return "int32"
	case strings.HasPrefix(t, "varchar"), strings.HasPrefix(t, "char"), strings.HasPrefix(t, "text"), strings.HasSuffix(t, "text"):
		return "string"
	case strings.HasPrefix(t, "datetime"), strings.HasPrefix(t, "timestamp"), strings.HasPrefix(t, "date"):
		file.GetImport(&astbasic.PkgBasic{
			ModPath: "time",
			Name:    "time",
		})
		return "time.Time"
	case strings.HasPrefix(t, "float"):
		return "float32"
	case strings.HasPrefix(t, "double"), strings.HasPrefix(t, "decimal"):
		return "float64"
	case strings.HasPrefix(t, "json"):
		return "interface{}"
	default:
		return "interface{}"
	}
}
