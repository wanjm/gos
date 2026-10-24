package callable_gen

import (
	"testing"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/astinfo"
	"github.com/wanjm/gos/basic"
)

func TestPermissionCheckCall(t *testing.T) {
	old := astinfo.GlobalProject
	t.Cleanup(func() { astinfo.GlobalProject = old })
	astinfo.GlobalProject = &astinfo.MainProject{Cfg: &basic.Config{}}
	astinfo.GlobalProject.Cfg.Generation.PermissionCheckFunc = "CheckPermission"
	astinfo.GlobalProject.Cfg.Generation.PermissionCheckMod = "gitlab.plaso.cn/nt/nt-common/filter"
	file := astbasic.CreateGenedFile("perm_test")
	method := &astinfo.Method{}
	method.Name = "CreateUser"
	if got := permissionCheckCall(method, file); got != "" {
		t.Fatalf("empty permission got %q", got)
	}
	method.Comment.Permission = "0,0"
	got := permissionCheckCall(method, file)
	want := "filter.CheckPermission(c, 0,0)"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}

	pkg := astinfo.NewPackage("example.com/mod/user_permission", false, "/tmp")
	pkg.Name = "user_permission"
	pkg.Enums = map[string]*astinfo.EnumDef{
		"PermissionGroup": {
			Members: []astinfo.EnumMember{{GoName: "PermissionGroupUser"}},
		},
		"PermissionType": {
			Members: []astinfo.EnumMember{{GoName: "PermissionTypeCreateAdmin"}},
		},
	}
	astinfo.GlobalProject.Packages = map[string]*astinfo.Package{pkg.ModPath: pkg}
	method.Comment.Permission = "user_permission.PermissionGroupUser,user_permission.PermissionTypeCreateAdmin"
	got = permissionCheckCall(method, file)
	want = "filter.CheckPermission(c, user_permission.PermissionGroupUser,user_permission.PermissionTypeCreateAdmin)"
	if got != want {
		t.Fatalf("named permission got %q want %q", got, want)
	}
}
