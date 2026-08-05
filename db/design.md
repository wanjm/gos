## mysql 说明
1. 根据给定的DB的DSN，获取数据库的ddl；
2. 给每个模块生成多个table；
3. 生成table时，自动生成dal；
### 配置文件说明
会自动在OutPath下生成dal目录和entity目录，dal会引用entity目录；package名字会同步使用这个相对目录
```
[[DBConfig]]
DSN="user:passwd@tcp(dbhost:3306)/dbplaso?charset=utf8mb4&parseTime=True&loc=Local"
DBName = "variable db name in dal,should be same as that in inititorator"
DBType = "mysql"

[[DBConfig.DbGenCfgs]]
OutPath = "business/package" # 生成的table所在的相对工程根目录
Tables = [
  { Name = "table_name", Arrays = ["column_name"], Maps = ["column_name"] }
]
```

### DefaultOrders（List / GetLimitAll 默认排序）

在 `project.public.toml` 的 `[Generation]` 中配置默认排序候选列表。生成 MySQL DAL 时，按列表顺序查找**当前表 entity 上存在的第一个列**，作为 `List` / `GetLimitAllWithStart` 的默认 `OrderFields`。

```toml
[Generation]
DefaultOrders = [
  { Field = "created_at", Direction = "DESC" },
  { Field = "create_time" },  # Direction 省略 → 复用第 0 项的 Direction
  { Field = "id" },
]
```

规则：
1. `Field`：DB 列名（与 entity 的 `DbColumnName` / gorm column 一致）。
2. `Direction`：`"ASC"` 或 `"DESC"`（大小写不敏感）。
3. 下标 `>= 1` 的项若省略 `Direction`，使用下标 `0` 的 `Direction`；下标 `0` 若也省略，则按 `DESC`。
4. 若未配置 `DefaultOrders`，默认等价于 `[{ Field = "create_time", Direction = "DESC" }, { Field = "id", Direction = "DESC" }]`。
5. 若列表中没有任何列存在于该表，则生成的查询不带默认 `OrderFields`。

生成调用链：
`List` / `GetLimitAllWithStart` → `ListWithOrder` / `GetLimitAllWithOrder`（`OrderByParams`）→ `ListWithOptions` / `GetLimitAllWithOptions`（`*SqlQueryOptions`）。
业务侧若需非默认排序，直接调用 `ListWithOrder` / `GetLimitAllWithOrder`。

## mongo 说明
1. 表明；
2. entity必须已经存在。表明和entity名字符合下划线到驼峰命名规则。
3. 生成的dal会引用entity目录；package名字会同步使用这个相对目录

## join table 说明
1. table; column, const;
2. entity -> schema;