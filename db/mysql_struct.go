package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/wanjm/gos/basic"
	"github.com/wanjm/gos/tool"
)

// GenTableFromMySQL connects to MySQL, gets the DDL of a table, and generates a Go struct definition.
func GenTableFromMySQL(config *basic.DBConfig, tableName string) error {
	if strings.ToLower(config.DBType) != "mysql" {
		return fmt.Errorf("当前仅支持 'mysql' 数据库类型, 您提供的是 '%s'", config.DBType)
	}

	// 1. Connect to MySQL
	db, err := connectToMySQL(config.DSN)
	if err != nil {
		return fmt.Errorf("无法连接到 MySQL: %w", err)
	}
	defer db.Close()
	log.Println("成功连接到 MySQL!")

	// 2. Get DDL
	ddl, err := getTableDDL(db, tableName)
	if err != nil {
		return fmt.Errorf("获取表 DDL 失败: %w", err)
	}
	log.Printf("表 %s 的 DDL: \n%s", tableName, ddl)

	// 3. Parse DDL and generate struct
	structCode, err := GenerateStructFromDDL(tableName, ddl)
	if err != nil {
		return fmt.Errorf("生成结构体代码失败: %w", err)
	}

	fmt.Println("生成的结构体定义:\n", structCode)
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
func GenerateStructFromDDL(tableName, ddl string) (string, error) {
	// Simple parser: extract column lines from DDL
	lines := strings.Split(ddl, "\n")
	var fields []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "`") {
			parts := strings.SplitN(line, " ", 3)
			if len(parts) < 2 {
				continue
			}
			colName := strings.Trim(parts[0], "`")
			colType := parts[1]
			goType := mysqlTypeToGoType(colType)
			fieldName := tool.ToPascalCase(colName, true)
			tag := fmt.Sprintf("`json:\"%s\" gorm:\"%s\"`", tool.FirstLower(colName), colName)
			fields = append(fields, fmt.Sprintf("    %s %s %s", fieldName, goType, tag))
		}
	}
	structName := tool.ToPascalCase(tableName, true)
	structDef := fmt.Sprintf("type %s struct {\n%s\n}", structName, strings.Join(fields, "\n"))
	return structDef, nil
}

// mysqlTypeToGoType maps MySQL types to Go types (basic mapping)
func mysqlTypeToGoType(mysqlType string) string {
	t := strings.ToLower(mysqlType)
	switch {
	case strings.HasPrefix(t, "int"):
		return "int"
	case strings.HasPrefix(t, "bigint"):
		return "int64"
	case strings.HasPrefix(t, "tinyint(1)"):
		return "bool"
	case strings.HasPrefix(t, "tinyint"):
		return "int8"
	case strings.HasPrefix(t, "smallint"):
		return "int16"
	case strings.HasPrefix(t, "mediumint"):
		return "int32"
	case strings.HasPrefix(t, "varchar"), strings.HasPrefix(t, "char"), strings.HasPrefix(t, "text"):
		return "string"
	case strings.HasPrefix(t, "datetime"), strings.HasPrefix(t, "timestamp"), strings.HasPrefix(t, "date"):
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
