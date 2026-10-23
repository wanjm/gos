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

func TestFolderOnServletStructComment(t *testing.T) {
	var c structComment
	ParseGosTags(`type="servlet" folder="直播带货"`, &c)
	if c.serverType != Servlet || c.Folder != "直播带货" {
		t.Fatalf("type=%q folder=%q", c.serverType, c.Folder)
	}
}

func TestHasApifoxFolderEnablesDirectoryUpdate(t *testing.T) {
	sw := NewSwagger(&MainProject{Cfg: &basic.Config{}})
	if sw.hasApifoxFolder() {
		t.Fatal("empty swagger should not update directories")
	}
	sw.addServletFromFunctionManager(&MethodManager{Server: []*Method{{
		Function: Function{
			FunctionField: FunctionField{Name: "List"},
			Comment:       functionComment{Url: "/tk/live/product/teacher/pendingList", Method: POST},
		},
		Receiver: &Struct{Comment: structComment{Folder: "直播带货"}},
	}}}, nil)
	if !sw.hasApifoxFolder() {
		t.Fatal("struct folder should mark the document for directory update")
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
		Tags []struct {
			Name string `json:"name"`
		} `json:"tags"`
		Paths map[string]map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	post := doc.Paths["/homework/list"]["post"]
	if got := post["x-apifox-folder"]; got != "homework" {
		t.Fatalf("x-apifox-folder = %v, want homework\n%s", got, raw)
	}
	tags, _ := post["tags"].([]any)
	if len(tags) != 1 || tags[0] != "homework" {
		t.Fatalf("tags = %v, want [homework]\n%s", post["tags"], raw)
	}
	if len(doc.Tags) != 1 || doc.Tags[0].Name != "homework" {
		t.Fatalf("root tags = %+v, want homework", doc.Tags)
	}
	ping := doc.Paths["/ping"]["post"]
	if _, ok := ping["x-apifox-folder"]; ok {
		t.Fatalf("servlet without folder should omit x-apifox-folder:\n%s", raw)
	}
	if pingTags, ok := ping["tags"].([]any); ok && len(pingTags) > 0 {
		t.Fatalf("servlet without folder should have empty tags: %v", pingTags)
	}
}
