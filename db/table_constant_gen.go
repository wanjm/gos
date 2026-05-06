package db

import (
	"fmt"
	"sort"
	"strings"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/basic"
)

// TryFetchTableConstantFromMySQLConfigs loads table_constant_config from the first
// reachable MySQL entry in the project config (used when generating Mongo entities).
func TryFetchTableConstantFromMySQLConfigs(cfgs []*basic.DBConfig) []tableConstantRow {
	for _, c := range cfgs {
		if strings.ToLower(c.DBType) != "mysql" || strings.TrimSpace(c.DSN) == "" {
			continue
		}
		dbConn, err := connectToMySQL(c.DSN)
		if err != nil {
			continue
		}
		rows, err := fetchAllTableConstantConfig(dbConn)
		_ = dbConn.Close()
		if err == nil {
			return rows
		}
	}
	return nil
}

// mongoFieldsColumnTypeMap maps BSON field name (table_constant_config.column_name) to Go type from generated struct fields.
func mongoFieldsColumnTypeMap(fields []FieldInfo) map[string]string {
	out := make(map[string]string, len(fields))
	for _, f := range fields {
		out[f.BsonTag] = f.Type
	}
	return out
}

func constLineComment(meaning string) string {
	s := strings.ReplaceAll(meaning, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}

// constUnderlyingType picks a numeric type suitable for integer constants typed as the entity field.
// Non-integer struct types fall back to int32.
func constUnderlyingType(fieldGoType string) string {
	t := strings.TrimPrefix(fieldGoType, "*")
	switch t {
	case "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64":
		return t
	default:
		return "int32"
	}
}

// genTableConstantFile writes const.gen.go using SQL/BSON column keys -> Go types from the entity struct.
func genTableConstantFile(tablepkg *astbasic.PkgBasic, columnToFieldType map[string]string, rows []tableConstantRow) {
	if len(rows) == 0 {
		return
	}
	constFile := tablepkg.NewFile("const")

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ColumnName != rows[j].ColumnName {
			return rows[i].ColumnName < rows[j].ColumnName
		}
		return rows[i].ConstName < rows[j].ConstName
	})

	colSeen := make(map[string]struct{})
	var cols []string
	for _, r := range rows {
		if _, ok := colSeen[r.ColumnName]; ok {
			continue
		}
		colSeen[r.ColumnName] = struct{}{}
		cols = append(cols, r.ColumnName)
	}
	sort.Strings(cols)

	var sb strings.Builder
	for _, col := range cols {
		ft := columnToFieldType[col]
		realType := constUnderlyingType(ft)
		typeName := astbasic.ToCamelCase(col, true) + "Type"
		sb.WriteString(fmt.Sprintf("\ntype %s %s\n", typeName, realType))
	}

	byCol := make(map[string][]tableConstantRow)
	for _, r := range rows {
		byCol[r.ColumnName] = append(byCol[r.ColumnName], r)
	}
	for _, col := range cols {
		rlist := byCol[col]
		typeName := astbasic.ToCamelCase(col, true) + "Type"
		sb.WriteString("\nconst (\n")
		for _, r := range rlist {
			cname := astbasic.ToCamelCase(r.ConstName, true)
			comment := constLineComment(r.Meaning)
			if comment != "" {
				sb.WriteString(fmt.Sprintf("\t%s %s = %d // %s\n", cname, typeName, r.Value, comment))
			} else {
				sb.WriteString(fmt.Sprintf("\t%s %s = %d\n", cname, typeName, r.Value))
			}
		}
		sb.WriteString(")\n")
	}

	constFile.AddBuilder(&sb)
	constFile.Save()
}
