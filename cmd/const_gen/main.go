package main

import (
	"log"
	"os"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/basic"
	"github.com/wanjm/gos/db"
)

const (
	sqlPath = "db/const_db.sql"
)

func main() {
	data, err := os.ReadFile(sqlPath)
	if err != nil {
		log.Fatalf("const_gen: read %s: %v (run from repository root)", sqlPath, err)
	}

	pkg := astbasic.SimplePackage("github.com/wanjm/gos/db", "db")
	pkg.FilePath = "db"
	genFile := pkg.NewFile("const_sql_struct")
	_, err = db.GenerateStructFromDDL("table_constant_config", string(data), genFile, "db", basic.TableCfg{})
	if err != nil {
		log.Fatalf("const_gen: %v", err)
	}
	genFile.Save()
}
