package astinfo

import "testing"

func TestParseSwaggerFlag(t *testing.T) {
	tests := []struct {
		in       string
		excluded bool
		ok       bool
	}{
		{"false", true, true},
		{"\"false\"", true, true},
		{"true", false, true},
		{"TRUE", false, true},
		{"", false, false},
		{"0", false, false},
		{"maybe", false, false},
	}
	for _, tt := range tests {
		excluded, ok := parseSwaggerFlag(tt.in)
		if excluded != tt.excluded || ok != tt.ok {
			t.Fatalf("parseSwaggerFlag(%q) = (%v, %v), want (%v, %v)", tt.in, excluded, ok, tt.excluded, tt.ok)
		}
	}
}

func TestSwaggerServletExcluded_annotation(t *testing.T) {
	servlet := &Method{
		Function: Function{
			FunctionField: FunctionField{Name: "Heartbeat"},
			Comment:       functionComment{Url: "/heartbeat", SwaggerExcluded: true},
		},
	}
	if !swaggerServletExcluded(servlet, nil) {
		t.Fatal("method swagger=false should be excluded")
	}
}

func TestSwaggerServletExcluded_structAnnotation(t *testing.T) {
	servlet := &Method{
		Function: Function{
			FunctionField: FunctionField{Name: "Ping"},
			Comment:       functionComment{Url: "/ping"},
		},
		Receiver: &Struct{
			Comment: structComment{SwaggerExcluded: true},
		},
	}
	if !swaggerServletExcluded(servlet, nil) {
		t.Fatal("struct swagger=false should exclude all methods")
	}
}

func TestSwaggerServletExcluded_globalPaths(t *testing.T) {
	servlet := &Method{
		Function: Function{
			FunctionField: FunctionField{Name: "Heartbeat"},
			Comment:       functionComment{Url: "/heartbeat"},
		},
	}
	if !swaggerServletExcluded(servlet, []string{"/heartbeat"}) {
		t.Fatal("ExcludePaths should exclude matching servlet url")
	}
	if swaggerServletExcluded(servlet, []string{"/api/heartbeat"}) {
		t.Fatal("ExcludePaths should not use UrlPrefix; match servlet url only")
	}
	if swaggerServletExcluded(servlet, []string{"/metrics"}) {
		t.Fatal("ExcludePaths should not match unrelated path")
	}
	servlet.Comment.Url = "/health/live"
	if !swaggerServletExcluded(servlet, []string{"/health/*"}) {
		t.Fatal("ExcludePaths wildcard should match servlet url")
	}
}
