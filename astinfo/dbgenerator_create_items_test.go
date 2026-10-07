package astinfo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wanjm/gos/astbasic"
)

func TestGenMysqlDalIncludesCreateItems(t *testing.T) {
	pkg := &astbasic.PkgBasic{Name: "dal", FilePath: t.TempDir()}
	file := pkg.NewFile("mysql.dal")
	entityPkg := astbasic.SimplePackage("example/entity/mysql/example_table", "example_table")
	data := &info{
		TableName:    "ExampleTable",
		RawTableName: "example_table",
		DBVariable:   "Mysql",
		Pkg:          entityPkg,
	}

	genMysqlDal(data, file)
	file.Save()

	content := readGeneratedFile(t, pkg.FilePath, "mysql.dal.gen.go")
	assertContains(t, content,
		"func (a *ExampleTableDal) CreateItems(ctx context.Context, items example_table.ExampleTableList) error",
		"err := dbOperation.Create(items)",
		`common.Error(ctx, "batch insert data to example_table failed", common.Err(err))`,
		"func (a *ExampleTableDal) Count(ctx context.Context, options []common.Optioner) (total int64, err error)",
		"QueryFields: options",
		"dbop.QueryCV(&common.SqlQueryOptions{",
		", &total, nil)",
		`common.Error(ctx, "Count record of example_table failed", common.Err(err))`,
	)
}

func TestGenMongoDalIncludesCreateItemsAndSetsIDs(t *testing.T) {
	pkg := &astbasic.PkgBasic{Name: "dal", FilePath: t.TempDir()}
	file := pkg.NewFile("mongo.dal")
	entityPkg := astbasic.SimplePackage("example/entity/mongo/example_table", "example_table")
	data := &info{
		TableName:    "ExampleTable",
		RawTableName: "example_table",
		DBVariable:   "Mongo",
		Pkg:          entityPkg,
		IDName:       "ID",
	}

	genMongoDal(data, file)
	file.Save()

	content := readGeneratedFile(t, pkg.FilePath, "mongo.dal.gen.go")
	assertContains(t, content,
		"func (a *ExampleTableDal) CreateItems(ctx context.Context, items example_table.ExampleTableList) error",
		"documents := make([]any, len(items))",
		"result, err := a.getDB().InsertMany(ctx, documents)",
		"items[i].ID = objectID",
		"func (a *ExampleTableDal) Count(ctx context.Context, options []common.Optioner) (total int64, err error)",
		"op := a.getOperation(ctx, options)",
		"total, err = op.Count()",
		`common.Error(ctx, "count example_table failed", common.Err(err))`,
	)
}

func readGeneratedFile(t *testing.T, dir, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func assertContains(t *testing.T, content string, values ...string) {
	t.Helper()
	for _, value := range values {
		if !strings.Contains(content, value) {
			t.Errorf("generated content does not contain %q", value)
		}
	}
}
