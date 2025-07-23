package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/wan_jm/servlet/astinfo"
	"github.com/wan_jm/servlet/astinfo/callable_gen"
)

func main() {
	var path string
	flag.StringVar(&path, "p", ".", "需要生成代码工程的根目录")
	init := flag.Bool("i", false, "初始化文件")
	h := flag.Bool("h", false, "显示帮助文件")
	flag.Parse()
	if *h {
		flag.Usage()
		return
	}
	path, err := filepath.Abs(path)
	if err != nil {
		log.Printf("open %s failed with %s", path, err.Error())
		return
	}
	os.Chdir(path)
	cfg := astinfo.Config{
		InitMain: *init,
	}
	// cfg.Load()
	astinfo.RegisterCallableGen(&callable_gen.ServletGen{}, &callable_gen.PrpcGen{}, &callable_gen.ResutfulGen{})
	var project = astinfo.CreateProject(path, &cfg)
	err = project.Parse()
	if err != nil {
		log.Printf("parse project failed with %s", err.Error())
		return
	}
	project.GenerateCode()
}
