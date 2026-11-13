package biz

import "gorm.io/gorm"

// Clock is constructed by gos and injected wherever *Clock is required.
// @gos auto
type Clock struct{}

// Worker shows field injection and wire:"-" on a second field of the same type.
// Clock is filled; Skip is left empty.
// @gos auto
type Worker struct {
	Clock *Clock
	Skip  *Clock `wire:"-"`
}

// Store shows same-type injection by matching field names to named initiator results.
// Db, Primary, and Replica each receive the *gorm.DB returned under that name.
// @gos auto
type Store struct {
	Db      *gorm.DB
	Primary *gorm.DB
	Replica *gorm.DB
}

// SharedClock receives the same *Clock instance that servlets inject.
// A package variable is assigned only when it is marked auto.
// @gos auto
var SharedClock *Clock
