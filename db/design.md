1. 根据给定的DB的DSN，获取数据库的ddl；
2. 给每个模块生成多个table；
3. 生成table时，自动生成dal；
4. 配置文件说明，会自动在OutPath下生成dal目录和entity目录，dal会引用entity目录；package名字会同步使用这个相对目录
配置文件示例
```
[[MysqlGenCfg]]
DSN="user:passwd@tcp(dbhost:3306)/dbplaso?charset=utf8mb4&parseTime=True&loc=Local"
DBName = "variable db name in dal,should be same as that in inititorator"

[[MysqlGenCfg.MysqlGenCfgs]]
OutPath = "business/package" # 生成的table所在的相对工程根目录
TableNames = ["table_name"]
```