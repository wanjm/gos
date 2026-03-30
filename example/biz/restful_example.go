package biz

import (
	"context"
	"errors"
)

// @gos type=restful;url="/api/v1/users"
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

// @gos url=""; method="POST"
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
