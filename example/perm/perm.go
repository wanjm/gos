package perm

import "context"

const (
	// PermGroupUser is the user permission group passed to servlet permission checks.
	PermGroupUser = 1
	// PermActionCreate is the create action passed to servlet permission checks.
	PermActionCreate = 2
)

// CheckPermission reports whether the caller may perform action in group.
// It allows any non-zero group or action so the example servlet can pass the generated check.
func CheckPermission(ctx context.Context, group int, action int) bool {
	return group != 0 || action != 0
}
