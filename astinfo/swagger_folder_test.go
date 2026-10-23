package astinfo

import (
	"encoding/json"
	"testing"

	"github.com/wanjm/gos/basic"
)

func TestSwaggerApifoxFolder(t *testing.T) {
	methodWins := &Method{
		Function: Function{Comment: functionComment{Folder: "homework"}},
		Receiver: &Struct{Comment: structComment{Folder: "order"}},
	}
	if got := swaggerApifoxFolder(methodWins); got != "homework" {
		t.Fatalf("method folder = %q, want homework", got)
	}

	structOnly := &Method{
		Receiver: &Struct{Comment: structComment{Folder: "order/mall"}},
	}
	if got := swaggerApifoxFolder(structOnly); got != "order/mall" {
		t.Fatalf("struct folder = %q, want order/mall", got)
	}

	if got := swaggerApifoxFolder(&Method{}); got != "" {
		t.Fatalf("empty folder = %q, want empty", got)
	}
	if swaggerApifoxFolder(nil) != "" {
		t.Fatal("nil servlet should have no folder")
	}
}

func TestFolderCommentParsed(t *testing.T) {
	var method functionComment
	method.dealValuePair("folder", `"order/mall"`)
	if method.Folder != "order/mall" {
		t.Fatalf("method folder = %q, want order/mall", method.Folder)
	}

	var class structComment
	class.dealValuePair("folder", `"homework"`)
	if class.Folder != "homework" {
		t.Fatalf("struct folder = %q, want homework", class.Folder)
	}
}

func TestAddServletWritesApifoxFolder(t *testing.T) {
	sw := NewSwagger(&MainProject{Cfg: &basic.Config{}})
	withFolder := &Method{
		Function: Function{
			FunctionField: FunctionField{Name: "List"},
			Comment:       functionComment{Url: "/homework/list", Title: "作业列表", Folder: "homework", Method: POST},
		},
	}
	withoutFolder := &Method{
		Function: Function{
			FunctionField: FunctionField{Name: "Ping"},
			Comment:       functionComment{Url: "/ping", Title: "ping", Method: POST},
		},
	}
	sw.addServletFromFunctionManager(&MethodManager{Server: []*Method{withFolder, withoutFolder}}, nil)
	raw, err := sw.swag.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Paths map[string]map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if got := doc.Paths["/homework/list"]["post"]["x-apifox-folder"]; got != "homework" {
		t.Fatalf("x-apifox-folder = %v, want homework\n%s", got, raw)
	}
	if _, ok := doc.Paths["/ping"]["post"]["x-apifox-folder"]; ok {
		t.Fatalf("servlet without folder should omit x-apifox-folder:\n%s", raw)
	}
}
