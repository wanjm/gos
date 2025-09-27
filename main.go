package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wanjm/gos/astinfo"
	"github.com/wanjm/gos/astinfo/callable_gen"
	rpcgen "github.com/wanjm/gos/astinfo/rpc_gen"
	"github.com/wanjm/gos/basic"
	"github.com/wanjm/gos/db"
)

func parseArgument() {
	flag.StringVar(&basic.Argument.SourcePath, "p", ".", "需要生成代码工程的根目录")
	flag.StringVar(&basic.Argument.ModName, "modname", "all", "指定模块名称")
	flag.StringVar(&basic.Argument.GoMod, "i", "", "本项目的gomod")
	flag.StringVar(&basic.Argument.SqlDBName, "dbname", "", "指定数据库名称")
	flag.StringVar(&basic.Argument.MongoDBName, "mongo", "", "指定Mongo数据库名称")
	flag.BoolVar(&basic.Argument.GenServlet, "s", false, "是否生成servlet代码")
	h := flag.Bool("h", false, "显示帮助文件")
	v := flag.Bool("v", false, "显示版本信息") // 添加-v参数
	flag.Parse()

	if *v { // 检查是否指定了-v参数
		fmt.Println("gos version 0.3.5") // 打印版本号
		return                           // 退出程序
	}

	if *h {
		flag.Usage()
		return
	}
	path, err := filepath.Abs(basic.Argument.SourcePath)
	if err != nil {
		fmt.Printf("open %s failed with %s", path, err.Error())
		return
	}
	basic.Argument.SourcePath = path
}

func main() {
	parseArgument()
	os.Chdir(basic.Argument.SourcePath)
	cfg := &basic.Cfg
	cfg.Load()
	var project = astinfo.CreateProject(basic.Argument.SourcePath, cfg)
	if err := project.CurrentProject.ParseModule(); err != nil {
		return
	}
	if basic.Argument.GenServlet {
		genServlet(project)
	}
	if basic.Argument.SqlDBName != "" {
		genDbData("mysql", db.GenTableForDb)
	}
	if basic.Argument.MongoDBName != "" {
		genDbData("mongo", db.GenTableForMongo)
	}
}
func genDbData(dbTypeName string, genDBFUntion func(config *basic.DBConfig, module string)) {
	var dbMap = make(map[string]*basic.DBConfig)
	var dbs = []string{}
	// 仅处理mysql，生成dbMap，和dbname数组
	for _, db := range basic.Cfg.DBConfig {
		if strings.ToLower(db.DBType) != dbTypeName {
			continue
		}
		dbMap[db.DBName] = db
		for _, module := range db.DbGenCfgs {
			module.ModulePath = astinfo.GlobalProject.CurrentProject.Module + "/" + module.OutPath
		}
		dbs = append(dbs, db.DBName)
	}
	// 如果为all，表示所有的db；
	var targetDbs []string
	if basic.Argument.SqlDBName == "all" {
		targetDbs = dbs
	} else {
		// 否则用逗号分隔的db；
		targetDbs = strings.Split(basic.Argument.SqlDBName, ",")
	}
	for _, dbName := range targetDbs {
		if cfg, ok := dbMap[dbName]; ok {
			genDBFUntion(cfg, dbName)
		} else {
			fmt.Printf("db %s not found", dbName)
		}
	}
}

func genServlet(project *astinfo.MainProject) {
	cfg := &basic.Cfg
	cfg.InitMain = basic.Argument.GoMod
	astinfo.RegisterCallableGen(
		callable_gen.NewServletGen(4, 1),
		&callable_gen.PrpcGen{},
		&callable_gen.ResutfulGen{},
		&callable_gen.RawGen{},
	)
	astinfo.RegisterClientGen(&rpcgen.PrpcGen{})

	// 移除原来的判断，因为现在InitMain直接存储模块名称
	// if len(modName) > 0 {
	// 	project.CurrentProject().Module = modName
	// }

	err := project.Parse()
	if err != nil {
		fmt.Printf("parse project failed with %s", err.Error())
		return
	}
	project.GenerateCode()
}
