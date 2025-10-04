package test

import (
	"testing"

	"github.com/wanjm/gos/tool"
)

func TestSplit(t *testing.T) {
	s := "json:\"country\" gorm:\"country\""
	a := tool.Fields(s)
	for _, v := range a {
		t.Log(v)
	}
}
