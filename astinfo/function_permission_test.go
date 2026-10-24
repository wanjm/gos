package astinfo

import "testing"

func TestPermissionCommentCopiedRaw(t *testing.T) {
	var c functionComment
	c.dealValuePair("permission", "0,1")
	if c.Permission != "0,1" {
		t.Fatalf("Permission=%q, want %q", c.Permission, "0,1")
	}
	c.dealValuePair("permission", `"0, 0"`)
	if c.Permission != "0, 0" {
		t.Fatalf("quoted Permission=%q, want %q", c.Permission, "0, 0")
	}
}
