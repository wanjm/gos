package callable_gen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wanjm/gos/astbasic"
)

// TestGenBasicErrorImportsJSON checks that generated prpc Error code can name json.RawMessage.
// It saves prpc.gen.go and requires an encoding/json import.
func TestGenBasicErrorImportsJSON(t *testing.T) {
	errorCodeGen = false
	t.Cleanup(func() { errorCodeGen = false })

	dir := t.TempDir()
	file := astbasic.CreateGenedFile("prpc")
	file.Pkg = &astbasic.PkgBasic{Name: "gen", FilePath: dir}
	GenBasicError(file)
	file.Save()

	src, err := os.ReadFile(filepath.Join(dir, "prpc.gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	if !strings.Contains(text, `"encoding/json"`) {
		t.Fatalf("missing encoding/json import:\n%s", text)
	}
	if !strings.Contains(text, "json.RawMessage") {
		t.Fatalf("missing json.RawMessage:\n%s", text)
	}
}
