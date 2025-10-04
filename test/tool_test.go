package test

import (
	"testing"

	"github.com/wanjm/gos/astbasic"
)

func TestSplit(t *testing.T) {
	s := "json:\"country\" gorm:\"country\""
	a := astbasic.Fields(s)
	for _, v := range a {
		t.Log(v)
	}
}
