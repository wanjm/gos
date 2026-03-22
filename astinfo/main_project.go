package astinfo

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"text/template"

	"github.com/wanjm/gos/astbasic"
	"github.com/wanjm/gos/basic"
	"golang.org/x/mod/modfile"
)

type MainProject struct {
	genPkg             *astbasic.PkgBasic // 记录gen包的信息
	CurrentProject     Project
	Packages           map[string]*Package // 项目包含的包集合（key为包全路径）
	SortedPacakgeNames []string
	Cfg                *basic.Config
	EntityMap          map[string]*Struct // entity structs by name (key: struct name, value: struct)

	*InitManager
	InitFuncs4All    []string   // 启动服务器和启动test都是用的方法；
	InitFuncs4Server []string   // 启动服务器用的方法；
	Projects         []*Project // 项目包含的子项目集合（key为Project的module）
}

func (mp *MainProject) genGoMod() {
	_, err := os.Stat("go.mod")
	if os.IsNotExist(err) {
		version := runtime.Version()
		parts := strings.Split(version, " ")
		//仅要版本号，去除
		version = parts[0]
		version = strings.Replace(version, "go", "go ", 1)
		var content = "module " + mp.Cfg.InitMain + "\n" + version + "\n"
		os.WriteFile("go.mod", []byte(content), 0660)
	}
}

// writeScaffoldFileIfAbsent creates relPath with content only if it does not exist (-i 初始化时不覆盖已有文件)。
func writeScaffoldFileIfAbsent(relPath string, content []byte) error {
	if _, err := os.Stat(relPath); err == nil {
		return nil
	}
	dir := filepath.Dir(relPath)
	if dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0750)
	}
	return os.WriteFile(relPath, content, 0660)
}

// genMain 与 lang_server 默认 main 一致（module 路径来自 -i）
func (mp *MainProject) genMain() {
	mod := mp.CurrentProject.ModPath
	body := fmt.Sprintf(`package main

import (
	"context"
	"time"

	"github.com/wanjm/common"
	"%s/basic"
	"%s/gen"
)

func main() {
	parseArgument()
	run()
}
func parseArgument() {
	basic.ParseArgument()
}
func run() {
	common.InitLogger()
	manager := common.GracefulManager
	shutdown := gen.Start(gen.Config{
		Cors:       true,
		Addr:       basic.Cfg.Server.ServerHost,
		ServerName: "servlet", // this is the name of group tag in comments;
	})
	manager.Go("http server shutdown monitor", func(ctx context.Context) {
		shutdown(ctx, 5*time.Second)
	})
	manager.Wait()
}
`, mod, mod)
	_ = writeScaffoldFileIfAbsent("main.go", []byte(body))
}

// genBasic 生成 basic 包默认文件（与 lang_server 一致）
func (mp *MainProject) genBasic() {
	_ = os.Mkdir("basic", 0750)

	_ = writeScaffoldFileIfAbsent("basic/config.go", []byte(`package basic

import (
	"flag"
	"fmt"
	"os"

	"github.com/wanjm/common"
)

type CommandOption struct {
	ConfigPath string
	Help       bool
}

type Server struct {
	ServerHost string
}

type Config struct {
	Server        Server
	DbConfig      common.DbConfig
	CommandOption CommandOption
	// Rpc Rpc
}

// type Rpc struct {
// 	CourseGoPrefix string
// }

var Cfg Config

func ParseArgument() {
	common.Cfg = &Cfg.DbConfig
	versionFlag := flag.Bool("v", false, "Print version information")
	flag.StringVar(&Cfg.CommandOption.ConfigPath, "d", "configs", "配置文件目录")
	flag.BoolVar(&Cfg.CommandOption.Help, "h", false, "帮助")
	flag.Parse()
	if Cfg.CommandOption.Help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	const Version = "0.0.2"
	if *versionFlag {
		fmt.Println("dataentry-go version", Version)
		os.Exit(0)
	}
	// common.LoadConfigFile(&Cfg, path.Join(Cfg.CommandOption.ConfigPath, "business.public.toml"))
	// fmt.Printf("Cfg: %+v\n", Cfg.Server.ServerHost)
}
`))

	_ = writeScaffoldFileIfAbsent("basic/const.go", []byte(`package basic

const (
	SUCCESS               = iota //成功
	SERVER_INTERNAL_ERROR        //服务器内部错误
	ACCESS_TOKEN_ERROR           //访问令牌错误
	_
	INPUT_DATA_ERROR     //输入数据错误
	NO_RECORD_FOUND      // 数据不存在
	RECORD_ALREADY_EXIST // 数据已存在
	_
	_
	CLIENT_VERSION_OLD // 客户版本太老
	FORMAT_ERROR       // 格式错误
)
`))

	_ = writeScaffoldFileIfAbsent("basic/message.go", []byte(`package basic

type Error struct {
	Code    int    "json:\"code\""
	Message string "json:\"message\""
}

func (error *Error) Error() string {
	return error.Message
}
func New(code int, msg string) error {
	res := &Error{
		Code:    code,
		Message: msg,
	}
	return res
}
func (error *Error) GetErrorCode() int {
	return error.Code
}
`))
}

func (mp *MainProject) genProjectPublicToml() {
	toml := `[Generation]
TraceKey="TraceIdstruct"
TraceKeyMod="github.com/wanjm/common"
ResponseKey="ResponseData"
ResponseMod="filter"
RpcLoggerKey="RpcLogger"
RpcLoggerMod="github.com/wanjm/common"
CommonMod="github.com/wanjm/common"
FlutterPath="../lang_client/lib/data/http"
ParseProjects = ["github/wanjm/common"]
## DBConfig配置（切片类型，使用[[ ]]表示数组元素）
##[[DBConfig]]
#DSN="user:passwd@tcp(dbhost:3306)/dbplaso in private file"
#DBName = "mysqlDB"
#DBType = "mysql"
#[[ DBConfig.DbGenCfgs ]]
#OutPath = "business/basic"
#Tables = [
#  { Name = "admin_org" },
#]
#[[DBConfig]]
#DBName = "mongoDB"
#DBType = "mongo"
`
	_ = writeScaffoldFileIfAbsent("project.public.toml", []byte(toml))
}

func (mp *MainProject) genProjectCode() {
	genPkg := mp.CurrentProject.NewPkgBasic("gen", "gen")
	mp.genPkg = genPkg
	file := genPkg.NewFile("goservlet_project")
	file.GetImport(astbasic.SimplePackage("github.com/gin-gonic/gin", "gin"))
	mp.genBasicCode(file)
	mp.genPrepare(file)
	file.Save()
}
func (mp *MainProject) SortDataForGen() {
	var pkgNames []string
	for _, pkg := range mp.Packages {
		if len(pkg.Initiator) > 0 || len(pkg.Structs) > 0 {
			pkgNames = append(pkgNames, pkg.ModPath)
			var strcutNames []string
			for _, class := range pkg.Structs {
				strcutNames = append(strcutNames, class.StructName)
			}
			sort.Strings(strcutNames)
			pkg.SortedStructNames = strcutNames
		}
	}
	sort.Strings(pkgNames)
	mp.SortedPacakgeNames = pkgNames
}
func (mp *MainProject) genPrepare(file *GenedFile) {
	mp.SortDataForGen()
	dbManager := DbManager{}
	dbManager.Gen()
	mp.InitInitorator()
	mp.InitManager.Generate(file)
	mp.InitManager.GenterateTestCode(file)

	sm := CreateServerManager()
	sm.Prepare()
	sm.Generate(file)

	cm := NewRpcClientManager()
	cm.Prepare()
	cm.Generate(file)

	// 定义模板字符串
	const prepareTemplate = `
// gened by mp.genPrepare
func Prepare() {
	//from mp.InitFuncs4All
{{range .InitFuncs4All}}	{{.}}()
{{end}}
	prepareVariable()
}

func prepare() {
	Prepare()
	//from mp.InitFuncs4Server
{{range .InitFuncs4Server}}	{{.}}()
{{end}}}
// gened by mp.genPrepare
`

	// 创建并解析模板
	tpl, err := template.New("prepare").Parse(prepareTemplate)
	if err != nil {
		panic("Failed to parse prepare template: " + err.Error())
	}

	// 渲染模板到strings.Builder
	var content strings.Builder
	if err := tpl.Execute(&content, mp); err != nil {
		panic("Failed to execute prepare template: " + err.Error())
	}

	file.AddBuilder(&content)
}
func (mp *MainProject) genBasicCode(file *GenedFile) {
	file.GetImport(astbasic.SimplePackage("github.com/gin-contrib/cors", "cors"))
	file.GetImport(astbasic.SimplePackage("sync", "sync"))
	file.GetImport(astbasic.SimplePackage("context", "context"))
	file.GetImport(astbasic.SimplePackage("net/http", "http"))
	file.GetImport(astbasic.SimplePackage("time", "time"))
	file.GetImport(astbasic.SimplePackage("fmt", "fmt"))

	var content strings.Builder
	content.WriteString(`
	type Response struct {
		Code     int    "json:\"code\""
		Message  string "json:\"message,omitempty\""
		ExtraInfo any   "json:\"extra,omitempty\"" //用于在失败的情况下也返回给前端一些信息；
		Object   any    "json:\"obj\""
		TraceId  string "json:\"traceId,omitempty\""
	}

type Config struct {
	CertFile string
	KeyFile string
	Cors bool
	Addr string
	ServerName string
	UrlDir     string // 静态文件的url目录, 默认为"/web"
	StaticDir  string // 静态文件目录, 为空时不开启；
}
func getAddr[T any](a T)*T{
	return &a
}
type server struct {
	filters      gin.HandlersChain
	routerInitors []func(*gin.Engine)
}
var servers map[string]*server
func Run(config ...Config) *sync.WaitGroup {
	prepare()
	var wg sync.WaitGroup
	for _, c := range config {
		wg.Add(1)
		go func() {
			defer wg.Done()
			srv := createServer(c)
			run(srv, c)
		}()
	}
	return &wg
}

// Start 返回一个函数，用于在上下文结束时关闭服务器
func Start(config ...Config) func(ctx context.Context, stopWaitTime time.Duration) {
	var servers []*http.Server
	prepare()
	for _, c := range config {
		srv := createServer(c)
		servers = append(servers, srv)
		go run(srv, c)
	}
	return func(ctx context.Context, stopWaitTime time.Duration) {
		// 这里我们其实是在等待 ctx.Done()，因为 manager.Go 内部调用的 fn 会立即执行
		// 但我们需要阻塞在这里等待信号
		<-ctx.Done()
		// 设定一个超时时间，强制结束未完成的请求（例如 5 秒）
		shutdownCtx, cancel := context.WithTimeout(context.Background(), stopWaitTime)
		defer cancel()
		for _, srv := range servers {
			if err := srv.Shutdown(shutdownCtx); err != nil {
				fmt.Printf("Server Shutdown Forced: %v\n", err)
			}
		}
	}
}

func run(srv *http.Server, config Config) {
	var err error
	if config.CertFile != "" {
		err = srv.ListenAndServeTLS(config.CertFile, config.KeyFile)
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil {
		fmt.Printf("listen: %v\n", err)
	}
}

func createServer(config Config) *http.Server {
	var router *gin.Engine = gin.New()
	router.ContextWithFallback = true
	if config.Cors {
		config := cors.DefaultConfig()
		config.AllowAllOrigins = true
		config.AllowHeaders = append(config.AllowHeaders, "*")
		router.Use(cors.New(config))
	}
	register(config.ServerName, router)
	if config.StaticDir != "" {
		url := config.UrlDir
		if url != "" {
			if url[0] != '/' {
				url = "/" + url
			}
		}else{
			url = "/web"
		}
		router.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, url)
		})	
		router.Static(url, config.StaticDir)
	}
	srv := &http.Server{
		Addr:    config.Addr,
		Handler: router,
	}
	return srv
}
		const TraceId = "TraceId"
	`)

	file.AddBuilder(&content)
}

type Server struct {
	Name             string
	callGen          CallableGen // 过去默认 一个group只能有一个functionType，但是考虑到位了节省端口号，有可能可以将多个functionType合并，然后再通过url去区分；目前仅降级为filter使用。类已经重新获取generator了。
	GeneratedFilters []string
	GenerateRouters  []string
	filters          []*Function
	routers          []*Struct
	manager          *ServerManager
	UrlList          []string            // all servlet URLs in this server
	RefFilterNames   map[string]struct{} // filter names explicitly referenced by methods
	// initRouteFuns []string           //initRoute 调用的init函数； 有package生成，生成路由代码时生成，一个package生成一个路由代码
	// urlFilters    map[string]*Filter //记录url过滤器函数,key是url, url是原始文件中的url，可能包含引号
	// initFuncs     []string           //initAll 调用的init函数；
}

// generate
func (sm *Server) Generate(file *GenedFile) {
	generator := sm.callGen
	generator.GenerateCommon(file)
	for _, function := range sm.filters {
		filterName := generator.GenFilterCode(function, file)
		if filterName == "" {
			continue
		}
		sm.GeneratedFilters = append(sm.GeneratedFilters, filterName)
	}
	for _, class := range sm.routers {
		//generate begin;
		sm.GenerateRouters = append(sm.GenerateRouters, sm.generateBegin(class, file))
		functionTYpe := class.Comment.serverType
		generator := sm.manager.generator[functionTYpe]
		if generator == nil {
			fmt.Printf("failed to found generator %s\n", functionTYpe)
			continue
		}
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
		class.Comment.GroupName,
		class.GoSource.Pkg.Name,
		class.StructName,
		"router",
	}, "_")
	var declare strings.Builder
	var receiver = Variable{
		Type: NewPointerType(class),
		Name: "receiver",
		Wire: true,
	}
	declare.WriteString("func " + name + "(engine *gin.Engine) {\n")
	declare.WriteString(receiver.Name + ":=" + receiver.Generate(file))
	declare.WriteString("\n")
	file.AddBuilder(&declare)
	file.GetImport(astbasic.SimplePackage("github.com/gin-gonic/gin", "gin"))
	return name
}

// collectFromRouter collects filter names and servlet URLs from the router into this server.
func (s *Server) collectFromRouter(router *Struct) {
	for _, method := range router.MethodManager.Server {
		for _, name := range strings.Split(method.Comment.Filter, ",") {
			name = strings.Trim(name, "\t ")
			if name != "" {
				s.RefFilterNames[name] = struct{}{}
			}
		}
		methodUrl := strings.Trim(method.Comment.Url, "\"")
		if methodUrl != "" {
			fullUrl := strings.Trim(path.Join(router.Comment.Url, methodUrl), "\"")
			s.UrlList = append(s.UrlList, fullUrl)
		}
	}
}

// filterNeeded returns true if the filter should be generated for this server.
func (s *Server) filterNeeded(filter *Function, pkgModPath string) bool {
	if s.manager.isMainProject(pkgModPath) {
		return true
	}
	if _, has := s.RefFilterNames[filter.Name]; has {
		return true
	}
	// If filterUrl is null/empty, the filter is needed
	if filter.Comment.Url == "" || filter.Comment.Url == "\"\"" {
		return true
	}
	filterUrl := strings.Trim(filter.Comment.Url, "\"")
	for _, methodUrl := range s.UrlList {
		if strings.Contains(methodUrl, filterUrl) {
			return true
		}
	}
	return false
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

// isMainProject returns true if the package belongs to the main (current) project.
func (sm *ServerManager) isMainProject(pkgModPath string) bool {
	mainMod := GlobalProject.CurrentProject.ModPath
	return pkgModPath == mainMod || strings.HasPrefix(pkgModPath, mainMod+"/")
}

// 扫描所有的程序，将服务按照group分为多个server；
func (sm *ServerManager) splitServers() {
	project := GlobalProject
	for _, pkgModuleName := range project.SortedPacakgeNames {
		pkg := project.Packages[pkgModuleName]
		// 结构体会定义group和type，所以先扫描struct
		for _, structName := range pkg.SortedStructNames {
			router := pkg.Structs[structName]
			var server *Server
			var ok bool
			var groupName = router.Comment.GroupName
			if groupName == "" {
				continue
			}
			// 过去默认一个group中只有一种服务类型；
			// 但是为了少开端口，现在一个group中可以有多种服务类型；
			// 此时有几个注意点；
			// 1. filter最好全部分开，这样每个filter在发现错误时可以正确返回对对应格式的报文；
			if server, ok = sm.servers[groupName]; !ok {
				gen := sm.generator[router.Comment.serverType]
				if gen == nil {
					fmt.Printf("failed to found generator %s\n", router.Comment.serverType)
					continue
				}
				server = &Server{
					Name:           groupName,
					callGen:        gen,
					manager:        sm,
					RefFilterNames: make(map[string]struct{}),
				}
				sm.servers[groupName] = server
			}
			server.routers = append(server.routers, router)
			server.collectFromRouter(router)
		}
	}
	for _, pkg := range project.Packages {
		for _, filter := range pkg.Filter {
			var server *Server
			var ok bool
			var groupName = filter.Comment.groupName
			if server, ok = sm.servers[groupName]; ok {
				if server.filterNeeded(filter, pkg.ModPath) {
					server.filters = append(server.filters, filter)
				}
			} else {
				fmt.Printf("failed to found server %s\n", groupName)
			}
		}
	}
}

// Generate
func (sm *ServerManager) Generate(file *GenedFile) {
	// if len(sm.servers) == 0 {
	// 	return
	// }
	var groupNames []string
	for name, server := range sm.servers {
		//一个server一个文件；
		file1 := CreateGenedFile(server.Name)
		server.Generate(file1)
		file1.Save()
		groupNames = append(groupNames, name)
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
		func getErrorCode(err error) (int, string) {
	if err == nil {
		return 0, ""
	}
	var errorCode int
	var errMessage = err.Error()
	if basicError,ok:=err.(Coder);ok{
		errorCode = basicError.GetErrorCode()
	}else{
		errorCode = 1
	}
	return errorCode, errMessage
}
type Coder interface {
	GetErrorCode() int
}
type ExtraInfo interface {
	GetExtraInfo() any
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
	slices.Sort(groupNames)
	for _, groupName := range groupNames {
		server := sm.servers[groupName]
		s = append(s, &ServerInfo{
			Name:        server.Name,
			FilterNames: strings.Join(server.GeneratedFilters, ",\n"),
			RouterNames: strings.Join(server.GenerateRouters, ",\n"),
		})
	}

	err = tmpl.Execute(&sb, s)
	if err != nil {
		log.Fatalf("执行模板失败: %v", err)
	}
	GlobalProject.InitFuncs4Server = append(GlobalProject.InitFuncs4Server, "initServer")
	file.AddBuilder(&sb)
}

// GetPackage retrieves a package by module path without creation
func (mp *MainProject) GetPackage(module string) *Package {
	return mp.Packages[module]
}

// FindPackage finds or creates a package with automatic module path resolution
func (mp *MainProject) FindPackage(module string) *Package {
	if pkg := mp.GetPackage(module); pkg != nil {
		return pkg
	}
	if module == mp.CurrentProject.ModPath+"/gen" {
		newPkg := NewPackage(module, true, filepath.Join(mp.CurrentProject.FilePath, "gen"))
		newPkg.finshedParse = true
		newPkg.Name = "gen"
		return newPkg
	}

	for _, p := range mp.Projects {
		// 根据module寻找package
		if p.ModPath == "" {
			panic(fmt.Sprintf("project module is empty %s\n", p.FilePath))
		}
		if strings.HasPrefix(module, p.ModPath) {
			//filepath.Join会换/\;
			newPkg := NewPackage(module, p.Simple, filepath.Join(p.FilePath, module[len(p.ModPath):]))
			mp.Packages[module] = newPkg
			newPkg.SimpleParse()
			return newPkg
		}
	}
	newPkg := NewSysPackage(module)
	newPkg.SimpleParse()
	mp.Packages[module] = newPkg
	//此处识别为系统Package
	return newPkg
}

// GenerateCode 生成项目的代码
func (mp *MainProject) GenerateCode() error {
	// 遍历所有包
	for _, pkg := range mp.Packages {
		_ = pkg
		// 生成包的代码
		// if err := pkg.GenerateCode(); err != nil {
		// 	return fmt.Errorf("error generating code for package %s: %w", pkg.Name, err)
		// }
	}
	if mp.Cfg.InitMain != "" {
		mp.genMain()
		mp.genBasic()
		mp.genProjectPublicToml()
	}
	mp.genProjectCode()

	NewSwagger(mp).GenerateCode(&mp.Cfg.SwaggerCfg)
	NewEntityGen(mp).GenerateCode()
	for _, gen := range projectGenerators {
		gen.GenerateCode(mp)
	}
	return nil
}

type ProjectGenerator interface {
	GenerateCode(mp *MainProject)
}

var projectGenerators []ProjectGenerator

func RegisterProjectGenerator(gen ...ProjectGenerator) {
	projectGenerators = append(projectGenerators, gen...)
}
func escapeModulePath(s string) string {
	var result strings.Builder
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result.WriteRune('!')
			result.WriteRune(r - 'A' + 'a') // 转为小写
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// resolveModulePath returns the filesystem path for a required module.
// It checks replace directives in the current project's go.mod; if a replace
// applies and points to a local path (New.Version == ""), that path is used.
// Otherwise falls back to GOPATH/pkg/mod.
func (mp *MainProject) resolveModulePath(mod *modfile.Require, goPath string) string {
	current := &mp.CurrentProject
	modPath := mod.Mod.Path
	version := mod.Mod.Version
	// Find longest matching replace (most specific wins)
	var best *modfile.Replace
	for _, r := range current.Replace {
		oldPath := r.Old.Path
		if modPath == oldPath || strings.HasPrefix(modPath, oldPath+"/") {
			if best == nil || len(oldPath) > len(best.Old.Path) {
				best = r
			}
		}
	}
	if best != nil && best.New.Version == "" {
		// Local path replacement
		replacePath := best.New.Path
		if filepath.IsAbs(replacePath) {
			return filepath.Clean(replacePath)
		}
		resolved := filepath.Join(current.FilePath, replacePath)
		abs, err := filepath.Abs(resolved)
		if err != nil {
			return filepath.Clean(resolved)
		}
		return abs
	}
	// Fallback to GOPATH
	return path.Join(goPath, "pkg/mod", escapeModulePath(modPath)+"@"+version)
}

func (mp *MainProject) ParseModule() error {
	return mp.CurrentProject.ParseModule()
}

// Parse 解析项目的代码
func (mp *MainProject) Parse() error {

	p := &mp.CurrentProject
	cfg := mp.Cfg
	traceKeyMod := cfg.Generation.TraceKeyMod
	if !strings.Contains(traceKeyMod, ".") {
		cfg.Generation.TraceKeyMod = p.ModPath + "/" + traceKeyMod
	}
	responseMod := cfg.Generation.ResponseMod
	if !strings.Contains(responseMod, ".") {
		cfg.Generation.ResponseMod = p.ModPath + "/" + responseMod
	}
	mp.Projects = append(mp.Projects, p)

	var projectsToParse []*Project
	projectsToParse = append(projectsToParse, p) // CurrentProject always first

	goPath := os.Getenv("GOPATH")
	parseProjectsSet := make(map[string]struct{})
	for _, modPath := range cfg.Generation.ParseProjects {
		parseProjectsSet[modPath] = struct{}{}
	}

	for _, mod := range p.Require {
		// if mod.Indirect {
		// 	continue
		// }

		filePath := mp.resolveModulePath(mod, goPath)
		proj := &Project{
			PkgBasic: astbasic.PkgBasic{
				FilePath: filePath,
			},
			Simple: true,
		}
		proj.ParseModule()
		if proj.ModPath == "" {
			proj.ModPath = mod.Mod.Path
		}
		mp.Projects = append(mp.Projects, proj)
		if _, ok := parseProjectsSet[mod.Mod.Path]; ok {
			proj.Simple = false // 需要扫描的project都不能simple扫描；
			projectsToParse = append(projectsToParse, proj)
		}
	}
	sort.Slice(mp.Projects, func(i, j int) bool {
		return mp.Projects[i].ModPath > mp.Projects[j].ModPath
	})

	for _, proj := range projectsToParse {
		if err := proj.ParseCode(); err != nil {
			return err
		}
	}
	mp.FinishedParse()
	return nil
}

// FinishedParse calls FinishedParse on each package for post-parse reordering.
func (mp *MainProject) FinishedParse() {
	for _, pkg := range mp.Packages {
		pkg.FinishedParse()
	}
}

var GlobalProject *MainProject

func CreateProject(path string, cfg *basic.Config) *MainProject {
	GlobalProject = &MainProject{
		Cfg:       cfg,
		Packages:  make(map[string]*Package),
		EntityMap: make(map[string]*Struct),
		// initiatorMap: make(map[*Struct]*Initiators),
		// servers:      make(map[string]*server),
		// creators: make(map[*Struct]*Initiator),
	}
	GlobalProject.CurrentProject.FilePath = path
	// 由于Package中有指向Project的指针，所以RawPackage指向了此处的project，如果返回对象，则出现了两个Project，一个是返回的Project，一个是RawPackage中的Project；
	// 返回*Project才能保证这是一个Project对象；
	// project.initRawPackage()
	// project.rawPkg = project.getPackage("", true)
	if cfg.InitMain != "" { // 检查是否非空字符串
		GlobalProject.genGoMod()
		// 设置项目模块名称
		GlobalProject.CurrentProject.ModPath = cfg.InitMain
	}
	return GlobalProject
}
