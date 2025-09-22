package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wanjm/gos/astinfo"
	"github.com/wanjm/gos/astinfo/callable_gen"
	rpcgen "github.com/wanjm/gos/astinfo/rpc_gen"
	"github.com/wanjm/gos/basic"
)

func parseArgument() {
	flag.StringVar(&basic.Argument.SourcePath, "p", ".", "需要生成代码工程的根目录")
	flag.StringVar(&basic.Argument.ModName, "i", "", "指定模块名称")
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
	genMysql()
}
func genMysql() {
	if basic.Argument.SqlPath == "" {
		return
	}
	var sqlMap = map[string]*basic.MysqlGenCfg{}
	for _, cfg := range basic.Cfg.MysqlGenCfgs {
		// sqlMap[cfg.ModulePath] = cfg
	}

}
func genServlet() {
	os.Chdir(basic.Argument.SourcePath)
	cfg := &basic.Cfg
	cfg.InitMain = basic.Argument.ModName
	cfg.Load()
	astinfo.RegisterCallableGen(
		callable_gen.NewServletGen(4, 1),
		&callable_gen.PrpcGen{},
		&callable_gen.ResutfulGen{},
		&callable_gen.RawGen{},
	)
	astinfo.RegisterClientGen(&rpcgen.PrpcGen{})
	var project = astinfo.CreateProject(basic.Argument.SourcePath, cfg)

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
