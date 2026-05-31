## 1. 创建工程
```
rm -fr server client
mkdir server client
cd server 
gos -i github.com/wanjm/gos_demo
go mod tidy
git init
git add .
git commit -m 'init version'
cd ../client
flutter create .
git init
git add .
git commit -m 'init version'
frc
cd ..
```


## 2. 添加服务端接口
```
cd server
mkdir -p business/biz business/schema
cat > business/schema/schema.go << 'EOF'
package schema

type HelloRequest struct {
	Name string `json:"name"` // 招呼名字
}

type HelloResponse struct {
	Message string `json:"message"` // 招呼内容
}
EOF


cat > business/biz/user.go << 'EOF'
package biz

import (
	"context"

	"github.com/wanjm/gos_demo/business/schema"
)

// @gos type="servlet"
type UserBiz struct {
}

// @gos url="/hello"; title="hello打招呼"
func (s *UserBiz) SayHello(ctx context.Context, req *schema.HelloRequest) (*schema.HelloResponse, error) {
	return &schema.HelloResponse{Message: "Hello, " + req.Name}, nil
}
EOF
git add .
git commit -m "add api hello"
```

## 3. 产生胶水代码,并提交
```
gos 
ga .
gc -m "add gen"
go run main.go
```
## 4. 验证服务端服务已经ready;
```
curl -X POST -H "Content-Type: application/json" -d '{"name":"wanjm"}' http://localhost:8080/hello
```

## 5. 验证前端生成代码
1. 打开project.public.toml 
2. 运行gos;


## 6. 配置前端环境pubspec.yaml
to dependencies:;
```
  component_set: ^0.1.2
  component_generator: ^0.1.1
  build_runner: ^2.4.0
``` 

to dev_dependencies:

``
flutter pub get
dart_gen
```

## 7. 修改前端代码
add to main.dart
```
  import 'data/http/network.gen.dart';
  import 'data/http/schema.gen.dart';

  String _counter = "";

  void _incrementCounter() async {
    var result = await userBizApi.sayHello(HelloRequest(name: "wanjm"));
    setState(() {
      if (result.code == 0) {
        _counter = result.obj!.message;
      }
    });
  }
```

## 8. apifox
[SwaggerCfg]
ProjectId=3497402
ServletFolder=87102439
SchemaFolder=19851892
Token=afxp_fb6b133IBkCNWKEO8se9OMuFhkoBROtA60PG


# example 2;
## 1. create sql;
```
mysql -u root -h localhost -pWjm_123456 << EOF
CREATE DATABASE IF NOT EXISTS demo;
USE demo;

CREATE TABLE student (
    id INT PRIMARY KEY AUTO_INCREMENT COMMENT '学生唯一标识ID',
    name VARCHAR(50) NOT NULL COMMENT '学生姓名',
    fullname VARCHAR(100) COMMENT '学生完整姓名',
    age TINYINT UNSIGNED COMMENT '学生年龄'
) ENGINE = INNODB 
DEFAULT CHARSET = utf8mb4 
COMMENT = '学生信息表';

INSERT INTO student (name, fullname, age) VALUES
('wanjm', 'wanjianming', 18),
('张三', '张三', 18),
('李四', '李四', 19),
('王五', '王五', 20),
('赵六', '赵六', 18),
('孙七', '孙七', 19),
('周八', '周八', 20),
('吴九', '吴九', 18),
('郑十', '郑十', 19),
('钱十一', '钱十一', 20),
('冯十二', '冯十二', 18);
EOF
```

## 2. 添加数据表
project.public.toml
```
[[DBConfig]]
DSN="root:Wjm_123456@tcp(localhost:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local"
DBName = "mysqlDB"
DBType = "mysql"
[[ DBConfig.DbGenCfgs ]]
OutPath = "business"
Tables = [
  { Name = "student" },
]
```
gos -dbname all

## 3. 配置应用数据库
business.public.toml
```
[DBConfig.MySqlConfig]
DSN="root:Wjm_123456@tcp(localhost:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local"
```
gos 产生DAL和ENTITY

## 4. 添加数据库逻辑
```
// @gos type="servlet"
type UserBiz struct {
	StudentDal *dal.StudentDal
}

// @gos url="/hello"; title="hello打招呼"
func (s *UserBiz) SayHello(ctx context.Context, req *schema.HelloRequest) (*schema.HelloResponse, error) {
	student, err := s.StudentDal.GetOne(ctx, common.OptEq(student.C_Name, req.Name))
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, errors.New("student not found")
	}
	return &schema.HelloResponse{Message: "Hello, " + student.Fullname}, nil
}
```

运行 gos 添加注入;

## 5. 
1. 刷新前端显示从数据库中得到的数据;