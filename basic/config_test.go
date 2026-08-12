package basic

import (
	"testing"

	"golang.org/x/mod/semver"
)

func TestToSemverAndCompare(t *testing.T) {
	cases := []struct {
		required string
		current  string
		wantGt   bool
	}{
		{"", "0.5.9", false},
		{"0.5.9", "0.5.9", false},
		{"0.5.8", "0.5.9", false},
		{"0.5.10", "0.5.9", true},
		{"0.6.0", "0.5.9", true},
		{"v0.5.9", "0.5.9", false},
	}
	for _, c := range cases {
		if c.required == "" {
			continue
		}
		req := toSemver(c.required)
		cur := toSemver(c.current)
		gotGt := semver.Compare(req, cur) > 0
		if gotGt != c.wantGt {
			t.Fatalf("required=%q current=%q gotGt=%v wantGt=%v", c.required, c.current, gotGt, c.wantGt)
		}
	}
}
