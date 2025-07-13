package main

import (
	pkg1 "github.com/wan_jm/servlet_example/pkg1"
)

type Student struct {
	Name   string
	Age    int
	Course []string
}

// @goservlet type=initiator
func GetSql1() *pkg1.StudentInPkg {
	return nil
}

// @goservlet type=initiator
func GetSql() *Student {
	return nil
}
