package db

import (
	"strings"

	"github.com/wanjm/gos/basic"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// 下划线转驼峰（首字母小写）
func toCamelCase(s string, useBig bool) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if !useBig {
			useBig = true
			continue // 保持第一个单词小写
		}
		// 后续单词首字母大写
		parts[i] = cases.Title(language.Dutch).String(part)
	}
	return strings.Join(parts, "")
}

// parseMethodName infers field name from method name. "UserIds" -> ("UserId", false), "UserIdMap" -> ("UserId", true).
func parseMethodName(methodName string) (fieldName string, isMap bool) {
	if strings.HasSuffix(methodName, "Map") {
		return methodName[:len(methodName)-3], true
	}
	if strings.HasSuffix(methodName, "s") && len(methodName) > 1 {
		return methodName[:len(methodName)-1], false
	}
	return "", false
}

// outputpath 输出路径，会自动输出到entity和dal目录下；
type MysqlGenCfg struct {
	Tables     []basic.TableCfg
	OutPath    string
	ModulePath string
}

// outputpath 输出路径，会自动输出到entity和dal目录下；
type MongoGenCfg = MysqlGenCfg
