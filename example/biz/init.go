package biz

import "gorm.io/gorm"

// @goservlet type=initiator
func GetSql() *HelloRequest {
	return nil
}

// @goservlet type=initiator
func GetSql2(a HelloRequest) *gorm.DB {
	return nil
}
