package biz

import (
	"context"
	"errors"
)

// UserRestful is the RESTful server. folder sets the Apifox directory for its routes.
// @gos type=restful; url="/api/v1/users"; folder="example/user"
type UserRestful struct {
}

type User struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type CreateUserReq struct {
	Name string `json:"name" binding:"required"`
	Age  int    `json:"age"`
}

// @gos url="/:id"; method="GET"
func (u *UserRestful) GetUser(ctx context.Context, req *User) (*User, error) {
	if req.Id == "1" {
		return &User{Id: "1", Name: "Alice", Age: 25}, nil
	}
	return nil, errors.New("user not found")
}

// CreateUser creates one user. Its folder overrides the struct directory.
// It returns the created user, or an error when the name is rejected.
// @gos url=""; method="POST"; title="create user"; folder="example/user/create"
func (u *UserRestful) CreateUser(ctx context.Context, req *CreateUserReq) (*User, error) {
	if req.Name == "error" {
		return nil, errors.New("invalid name")
	}
	return &User{Id: "2", Name: req.Name, Age: req.Age}, nil
}

// @gos url="/hello/:name"; method="GET"
func (u *UserRestful) SayHelloString(ctx context.Context, req *User) (string, error) {
	if req.Name == "error" {
		return "", errors.New("cannot say hello to error")
	}
	return "Hello, " + req.Name, nil
}

type PatchReq struct {
	Name   string `json:"name"`
	Source string `json:"source" header:"X-Source"`
}

// UpdateUser replaces one user. method=PUT selects the HTTP verb.
// It returns the updated user.
// @gos url="/:id"; method="PUT"; title="update user"
func (u *UserRestful) UpdateUser(ctx context.Context, req *User) (*User, error) {
	return &User{Id: req.Id, Name: req.Name, Age: req.Age}, nil
}

// DeleteUser removes one user. method=DELETE selects the HTTP verb.
// It returns the deleted id.
// @gos url="/:id"; method="DELETE"; title="delete user"
func (u *UserRestful) DeleteUser(ctx context.Context, req *User) (*User, error) {
	return &User{Id: req.Id}, nil
}

// PatchUser updates part of a user and reads X-Source from the header tag.
// swagger=false keeps this route out of the OpenAPI document.
// @gos url="/search"; method="PATCH"; title="patch user"; swagger=false
func (u *UserRestful) PatchUser(ctx context.Context, req *PatchReq) (*User, error) {
	return &User{Name: req.Name}, nil
}
