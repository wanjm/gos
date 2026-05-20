package astinfo

import (
	"encoding/json"
	"testing"
)

func TestSwaggerHeaderParameters_marshalsOpenAPI2Header(t *testing.T) {
	ps := swaggerHeaderParameters([]FieldBasic{{Name: "X-Test-Id"}})
	if len(ps) != 1 {
		t.Fatalf("len=%d", len(ps))
	}
	raw, err := json.Marshal(ps[0])
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["in"] != "header" || m["name"] != "X-Test-Id" || m["type"] != "string" || m["required"] != true {
		t.Fatalf("unexpected JSON: %s", string(raw))
	}
}

func TestSwaggerHeaderParameters_descriptionFromComment(t *testing.T) {
	ps := swaggerHeaderParameters([]FieldBasic{{
		Name:    "X-Trace",
		Comment: FieldComment{CommentText: "  correlation id  "},
	}})
	if len(ps) != 1 || ps[0].Description != "correlation id" {
		t.Fatalf("got %#v", ps[0])
	}
}

func TestSwaggerApplicableRouteFilters_explicitAndURL(t *testing.T) {
	withURL := &Function{
		FunctionField: FunctionField{Name: "WithURL"},
		Comment: functionComment{
			Url: "/api",
		},
	}
	named := &Function{
		FunctionField: FunctionField{Name: "Named"},
		Comment: functionComment{
			Url: "/x",
		},
	}
	servlet := &Method{
		Function: Function{
			FunctionField: FunctionField{Name: "M"},
			Comment: functionComment{
				Url:     "/api/foo",
				Filters: []string{"Named"},
			},
		},
	}
	got := swaggerApplicableRouteFilters([]*Function{withURL, named}, servlet)
	if len(got) != 2 || got[0] != named || got[1] != withURL {
		t.Fatalf("got %#v", got)
	}
}

func TestCollectServletSwaggerHeaders_fromFiltersDedupes(t *testing.T) {
	flt := &Function{
		FunctionField: FunctionField{Name: "F"},
		Comment: functionComment{
			Url: "/hello",
			RequiredHeaders: []FieldBasic{
				{Name: "X-A"},
				{Name: "x-a"},
				{Name: "X-B", Comment: FieldComment{CommentText: "second"}},
			},
		},
	}
	servlet := &Method{
		Function: Function{
			FunctionField: FunctionField{Name: "SayHello"},
			Comment:       functionComment{Url: "/hello"},
		},
		Receiver: &Struct{
			Comment: structComment{GroupName: "g"},
		},
	}
	byGroup := map[string][]*Function{"g": {flt}}
	got := collectServletSwaggerHeaders(servlet, byGroup)
	if len(got) != 2 || got[0].Name != "X-A" || got[1].Name != "X-B" || got[1].Comment.CommentText != "second" {
		t.Fatalf("got %#v", got)
	}
}
