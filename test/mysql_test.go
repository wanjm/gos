package test

import (
	"strings"
	"testing"
)

func TestMySQLStructGeneration(t *testing.T) {
	ddl := strings.ReplaceAll(`CREATE TABLE "student" (
  "id" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "loginname" char(190) NOT NULL COMMENT '登录名',
  "password" char(255) NOT NULL COMMENT '密码',
  "mCode" varchar(6) DEFAULT NULL COMMENT '电话的国际地区码',
  "mobile" varchar(45) DEFAULT NULL COMMENT '手机号',
  "name" varchar(128) DEFAULT NULL COMMENT '姓名',
  "country" varchar(128) DEFAULT 'CN' COMMENT '国家',
  "state" varchar(128) DEFAULT NULL COMMENT '地区',
  "city" varchar(128) DEFAULT NULL COMMENT '城市',
  "school" varchar(128) DEFAULT NULL COMMENT '学校',
  "signature" text COMMENT '签名',
  "profile" text COMMENT '年级信息',
  "avatarUrl" varchar(512) DEFAULT NULL COMMENT '头像',
  "createAt" datetime DEFAULT NULL COMMENT '创建日期',
  "createId" bigint(20) NOT NULL DEFAULT '-1' COMMENT '创建管理员 学生自己注册/历史数据为-1',
  "updateAt" datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日期',
  "userLevel" int(10) unsigned NOT NULL DEFAULT '1' COMMENT '用户等级',
  "active" int(10) unsigned NOT NULL DEFAULT '1' COMMENT '删除状态',
  "email" varchar(190) DEFAULT NULL COMMENT '邮箱',
  "studentNumber" varchar(128) DEFAULT NULL COMMENT '学号',
  "activeUntil" datetime DEFAULT NULL COMMENT '过期时间',
  "lastLogin" datetime NOT NULL DEFAULT '1980-01-28 00:00:00' COMMENT '最近登录',
  "validTimeInterval" int(11) DEFAULT '10',
  "org_id" int(10) unsigned DEFAULT NULL COMMENT '机构id',
  "sex" tinyint(4) DEFAULT '0' COMMENT '性别 0-其他 1-男 2-女',
  "bz" varchar(255) DEFAULT NULL COMMENT '备注',
  "register_from" int(1) NOT NULL DEFAULT '1' COMMENT '来源：1 机构注册  2  体验开放注册',
  "weixin_id" char(28) DEFAULT NULL COMMENT '微信ID',
  "real_class_id" int(8) DEFAULT NULL COMMENT 'Class ID in Mongo',
  "grade" int(11) DEFAULT NULL COMMENT '年级',
  "myFile" varchar(30) DEFAULT NULL COMMENT '我的文件目录ID',
  "myFav" varchar(30) DEFAULT NULL COMMENT '我的收藏目录ID',
  "myMini" varchar(30) DEFAULT NULL COMMENT '我的微课目录ID',
  "loginMobile" char(11) DEFAULT NULL COMMENT '登录手机号',
  "formalFlag" tinyint(4) NOT NULL DEFAULT '1' COMMENT '是否正式学生',
  "service_flag" tinyint(1) NOT NULL DEFAULT '1' COMMENT '账号类型 0:临时账号 1:服务账号 2:普通账号',
  "plasoCoin" int(11) NOT NULL DEFAULT '0' COMMENT '用户剩余的T币',
  "score" int(8) NOT NULL DEFAULT '0' COMMENT '总学分',
  "externalId" varchar(128) DEFAULT NULL COMMENT '关联外部系统的ID',
  "externalLoginName" varchar(128) DEFAULT NULL COMMENT '关联外部系统的loginName',
  "supportParent" tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否支持家长登录,默认为0不支持',
  "s0" varchar(255) DEFAULT NULL COMMENT '自定义字段1',
  "s1" varchar(255) DEFAULT NULL COMMENT '自定义字段2',
  "s2" varchar(255) DEFAULT NULL COMMENT '自定义字段3',
  "s3" varchar(255) DEFAULT NULL COMMENT '自定义字段4',
  "s4" varchar(255) DEFAULT NULL COMMENT '自定义字段5',
  "totalGetScore" int(11) DEFAULT '0' COMMENT '累计获得积分',
  "totaloutScore" int(11) DEFAULT '0' COMMENT '累计获得积分',
  "subscribeFlag" tinyint(1) DEFAULT '0' COMMENT '微信是否已关注 0：没有 1：关注',
  "followFlag" tinyint(1) DEFAULT '0' COMMENT '是否跟进 0：没有 1：已跟进',
  "hasPopQr" tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否弹过课程顾问/服务群',
  "ai_question_perday" int(11) NOT NULL DEFAULT '0',
  "deleted_time" int(10) NOT NULL DEFAULT '0' COMMENT '删除时间(在active=0下有效',
  PRIMARY KEY ("id"),
  UNIQUE KEY "loginname_UNIQUE" ("loginname"),
  KEY "weixin_id_idx" ("weixin_id"),
  KEY "mobile" ("loginMobile"),
  KEY "idx_externalName" ("externalLoginName"),
  KEY "idx_externalId" ("externalId"),
  KEY "idx_email" ("email"),
  KEY "idx_updateAt" ("updateAt"),
  KEY "idx_orgid_totalGetScore" ("org_id","totalGetScore"),
  KEY "idx_orgid_formalFlag_active_activeUntil_serviceFlag" ("org_id","formalFlag","active","activeUntil","service_flag") USING BTREE,
  KEY "idx_org_formal_active_loginname" ("org_id","formalFlag","active","loginname")
) ENGINE=InnoDB AUTO_INCREMENT=9935188 DEFAULT CHARSET=utf8mb4 COMMENT='机构学生表'
 `, `"`, "`")
	_ = ddl
	// content, err := db.GenerateStructFromDDL("student", ddl)
	// if err != nil {
	// 	t.Fatalf("Error generating struct: %v", err)
	// }
	// t.Logf("Generated struct:\n%s", content)
	// 这里可以添加测试代码来调用 GenTableFromMySQL 函数
}
