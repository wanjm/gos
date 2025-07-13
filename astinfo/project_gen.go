package astinfo

import (
	"log"
	"os"
	"runtime"
	"strings"
)

func (project *Project) genGoMod() {
	_, err := os.Stat("go.mod")
	if os.IsNotExist(err) {
		var content = "module " + project.Module + "\n" + strings.Replace(runtime.Version(), "go", "go ", 1) + "\n"
		os.WriteFile("go.mod", []byte(content), 0660)
	}
}

// genMain
func (project *Project) genMain() {
	var content strings.Builder
	content.WriteString("package main\n")
	//	import "gitlab.plaso.cn/message-center/gen"
	content.WriteString("import (\"flag\"\n\"" + project.Module + "/gen\")\n")
	content.WriteString(`
func main() {
	parseArgument();
	run()
}
func parseArgument() {
	flag.Parse()
}
func run() {
	wg:=gen.Run(gen.Config{
		Cors: true,
		Addr: ":8080",
		ServerName: "servlet", // this is the name of group tag in comments;
	})
	wg.Wait()
}
	`)
	os.WriteFile("main.go", []byte(content.String()), 0660)

}

// genBasic 生成basic.go
func (project *Project) genBasic() {
	os.Mkdir("basic", 0750)
	os.WriteFile("basic/message.go", []byte(`package basic
type Error struct {
	Code    int    "json:\"code\""
	Message string "json:\"message\""
}

func (error *Error) Error() string {
	return error.Message
}
	`), 0660)
}

func (project *Project) genInitMain() {
	//如果是空目录，或者init为true；则生成main.go 和basic.go的Error类；
	if !project.cfg.InitMain {
		return
	}
	project.genGoMod()
	project.genMain()
	project.genBasic()
}
func (p *Project) genProjectCode() {
	err := os.Mkdir("gen", 0750)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	file := createGenedFile("goservlet_project", p)
	file.getImport(SimplePackage("github.com/gin-gonic/gin", "gin"))
	os.Chdir("gen")
	p.genBasicCode(file)
	p.genPrepare(file)
	file.save()
}
func (p *Project) genPrepare(file *GenedFile) {
	p.InitInitorator()
	var content strings.Builder
	p.InitManager.Generate(file)
	content.WriteString("func Prepare() {\n")
	for _, fun := range p.initFuncs {
		content.WriteString(fun + "()\n")
	}
	content.WriteString(`
	}	
	func prepare() {
		Prepare()
    }
	`)
	file.addBuilder(&content)
}
func (Project *Project) genBasicCode(file *GenedFile) {
	file.getImport(SimplePackage("github.com/gin-contrib/cors", "cors"))
	file.getImport(SimplePackage("sync", "sync"))

	var content strings.Builder
	content.WriteString(`
	type Response struct {
		Code    int         "json:\"code\""
		Message string      "json:\"message,omitempty\""
		Object  interface{} "json:\"obj,omitempty\""
	}

type Config struct {
	CertFile string
	KeyFile string
	Cors bool
	Addr string
	ServerName string
}
func getAddr[T any](a T)*T{
	return &a
}
type server struct {
	filters      gin.HandlersChain
	routerInitors []func(*gin.Engine)
}
var servers map[string]*server
	func Run(config ...Config) *sync.WaitGroup{
		prepare()
		var wg sync.WaitGroup
		for _, c := range config {
			wg.Add(1)
			go run(&wg, c)
		}
		return &wg
	}
	func run(wg *sync.WaitGroup, config Config){
		var	router  *gin.Engine = gin.New()
		router.ContextWithFallback = true
		if(config.Cors){
			config := cors.DefaultConfig()
			config.AllowAllOrigins = true
			config.AllowHeaders = append(config.AllowHeaders, "*")
			router.Use(cors.New(config))
		}
			//如果不存在，则启动就失败，不需要检查
		server := servers[config.ServerName]
		if server.filters != nil {
			router.Use(server.filters...)
		}
		for _, routerInitor := range server.routerInitors {
			routerInitor(router)
		}
		if config.CertFile != "" {
			router.RunTLS(config.Addr, config.CertFile, config.KeyFile)
		} else {
			router.Run(config.Addr)
		}
		wg.Done()
	}
		const TraceId = "TraceId"
	`)

	file.addBuilder(&content)
}
