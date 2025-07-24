package astinfo

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"text/template"
)

type MainProject struct {
	Project
	Packages map[string]*Package // 项目包含的包集合（key为包全路径）
	Cfg      *Config

	*InitManager
	initFuncs []string
}

func (project *MainProject) genGoMod() {
	_, err := os.Stat("go.mod")
	if os.IsNotExist(err) {
		var content = "module " + project.Module + "\n" + strings.Replace(runtime.Version(), "go", "go ", 1) + "\n"
		os.WriteFile("go.mod", []byte(content), 0660)
	}
}

// genMain
func (project *MainProject) genMain() {
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
func (project *MainProject) genBasic() {
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

func (project *MainProject) genInitMain() {
	//如果是空目录，或者init为true；则生成main.go 和basic.go的Error类；
	if !project.Cfg.InitMain {
		return
	}
	project.genGoMod()
	project.genMain()
	project.genBasic()
}
func (p *MainProject) genProjectCode() {
	err := os.Mkdir("gen", 0750)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	file := createGenedFile("goservlet_project")
	file.GetImport(SimplePackage("github.com/gin-gonic/gin", "gin"))
	os.Chdir("gen")
	p.genBasicCode(file)
	p.genPrepare(file)
	file.save()
}
func (p *MainProject) genPrepare(file *GenedFile) {

	p.InitInitorator()
	p.InitManager.Generate(file)

	sm := CreateServerManager()
	sm.Prepare()
	sm.Generate(file)

	var content strings.Builder
	content.WriteString("func Prepare() {\n")
	for _, fun := range p.initFuncs {
		content.WriteString(fun + "()\n")
	}
	content.WriteString(`
	}	
	func prepare() {
		Prepare()
		initServer()
    }
	`)
	file.AddBuilder(&content)
}
func (MainProject *MainProject) genBasicCode(file *GenedFile) {
	file.GetImport(SimplePackage("github.com/gin-contrib/cors", "cors"))
	file.GetImport(SimplePackage("sync", "sync"))

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
		register(config.ServerName, router)
		if config.CertFile != "" {
			router.RunTLS(config.Addr, config.CertFile, config.KeyFile)
		} else {
			router.Run(config.Addr)
		}
		wg.Done()
	}
		const TraceId = "TraceId"
	`)

	file.AddBuilder(&content)
}

type Server struct {
	Name             string
	callGen          CallableGen
	GeneratedFilters []string
	GenerateRouters  []string
	filters          []*Function
	routers          []*Struct
	// initRouteFuns []string           //initRoute 调用的init函数； 有package生成，生成路由代码时生成，一个package生成一个路由代码
	// urlFilters    map[string]*Filter //记录url过滤器函数,key是url, url是原始文件中的url，可能包含引号
	// initFuncs     []string           //initAll 调用的init函数；
}

// generate
func (sm *Server) Generate(file *GenedFile) {
	generator := sm.callGen
	generator.GenerateCommon(file)
	for _, function := range sm.filters {
		sm.GeneratedFilters = append(sm.GeneratedFilters, generator.GenFilterCode(function, file))
	}
	for _, class := range sm.routers {
		//generate begin;
		sm.GenerateRouters = append(sm.GenerateRouters, sm.generateBegin(class, file))

		// generate servlets;
		for _, method := range class.MethodManager.Server {
			generator.GenRouterCode(method, file)
		}

		// generate end
		var end strings.Builder
		end.WriteString("}\n")
		file.AddBuilder(&end)
	}
}

// generateBegin
func (sm *Server) generateBegin(class *Struct, file *GenedFile) string {
	var name = strings.Join([]string{
		"init",
		class.comment.groupName,
		class.Pkg.Name,
		class.StructName,
		"router",
	}, "_")
	var declare strings.Builder
	var receiver = Variable{
		Type: class,
		Name: "receiver",
		Wire: true,
	}
	declare.WriteString("func " + name + "(engine *gin.Engine) {\n")
	declare.WriteString(receiver.Name + ":=" + receiver.Generate(file))
	declare.WriteString("\n")
	file.AddBuilder(&declare)
	file.GetImport(SimplePackage("github.com/gin-gonic/gin", "gin"))
	return name
}

// 负责对配置的每个server进行初始化，管理其中的filter，servlet；并生成最终代码中的server代码。打通filter和servlet的注册环节
// 其生成代码分为连个部分；
// 1. 最终代码的server代码。完成代码的filter和路由的注册；
// 2. filter，和路由的工作代码
type ServerManager struct {
	servers   map[string]*Server
	generator map[string]CallableGen
}

func CreateServerManager() *ServerManager {
	sm := &ServerManager{
		servers:   map[string]*Server{},
		generator: map[string]CallableGen{},
	}
	return sm
}

// register
func (sm *ServerManager) register(callGen CallableGen) {
	name := callGen.GetName()
	sm.generator[name] = callGen
}

func (sm *ServerManager) Prepare() {
	for _, callGen := range callableGens {
		sm.register(callGen)
	}
	sm.splitServers()
}

// 扫描所有的程序，将服务按照group分为多个server；
func (sm *ServerManager) splitServers() {
	project := GlobalProject
	for _, pkg := range project.Packages {
		// 结构体会定义group和type，所以先扫描struct
		for _, router := range pkg.Structs {
			var server *Server
			var ok bool
			var groupName = router.comment.groupName
			if groupName == "" {
				continue
			}
			if server, ok = sm.servers[groupName]; !ok {
				server = &Server{
					Name:    groupName,
					callGen: sm.generator[router.comment.serverType],
				}
				sm.servers[groupName] = server
			}
			server.routers = append(server.routers, router)
		}
		for _, filter := range pkg.Filter {
			var server *Server
			var ok bool
			var groupName = filter.Comment.groupName
			if server, ok = sm.servers[groupName]; ok {
				server.filters = append(server.filters, filter)
			} else {
				fmt.Printf("failed to found server %s", groupName)
			}
		}
	}
}

// Generate
func (sm *ServerManager) Generate(file *GenedFile) {
	for _, server := range sm.servers {
		//一个server一个文件；
		file1 := createGenedFile(server.Name)
		server.Generate(file1)
		file1.save()
	}
	tmplText :=
		`
func initServer(){
	servers = make(map[string]*server)
	{{range .}}
	servers["{{.Name}}"] = &server{
		filters: gin.HandlersChain{	{{.FilterNames}} },
		routerInitors: []func(*gin.Engine){ {{.RouterNames}} },
	}
	{{end}}
}

	func register(name string, router *gin.Engine ){
		server := servers[name]
		if server.filters != nil {
			router.Use(server.filters...)
		}
		for _, routerInitor := range server.routerInitors {
			routerInitor(router)
		}
	}
`
	tmpl, err := template.New("personInfo").Parse(tmplText)
	if err != nil {
		log.Fatalf("解析模板失败: %v", err)
	}
	var sb strings.Builder
	type ServerInfo struct {
		Name        string
		FilterNames string
		RouterNames string
	}
	var s []*ServerInfo
	for _, server := range sm.servers {
		server := &ServerInfo{
			Name:        server.Name,
			FilterNames: strings.Join(server.GeneratedFilters, ",\n"),
			RouterNames: strings.Join(server.GenerateRouters, ",\n"),
		}
		s = append(s, server)
	}

	err = tmpl.Execute(&sb, s)
	if err != nil {
		log.Fatalf("执行模板失败: %v", err)
	}

	file.AddBuilder(&sb)
}

// GetPackage retrieves a package by module path without creation
func (p *MainProject) GetPackage(module string) *Package {
	return p.Packages[module]
}

// FindPackage finds or creates a package with automatic module path resolution
func (p *MainProject) FindPackage(module string) *Package {
	if pkg := p.GetPackage(module); pkg != nil {
		return pkg
	}
	newPkg := NewPackage(module)
	p.Packages[module] = newPkg
	return newPkg
}

// GenerateCode 生成项目的代码
func (p *MainProject) GenerateCode() error {
	p.genInitMain()
	// 遍历所有包
	for _, pkg := range p.Packages {
		_ = pkg
		// 生成包的代码
		// if err := pkg.GenerateCode(); err != nil {
		// 	return fmt.Errorf("error generating code for package %s: %w", pkg.Name, err)
		// }
	}
	p.genProjectCode()
	return nil
}

var GlobalProject *MainProject

func CreateProject(path string, cfg *Config) *MainProject {
	GlobalProject = &MainProject{
		Cfg:      cfg,
		Packages: make(map[string]*Package),
		// initiatorMap: make(map[*Struct]*Initiators),
		// servers:      make(map[string]*server),
		// creators: make(map[*Struct]*Initiator),
	}
	GlobalProject.Path = path
	// 由于Package中有指向Project的指针，所以RawPackage指向了此处的project，如果返回对象，则出现了两个Project，一个是返回的Project，一个是RawPackage中的Project；
	// 返回*Project才能保证这是一个Project对象；
	// project.initRawPackage()
	// project.rawPkg = project.getPackage("", true)
	return GlobalProject
}
