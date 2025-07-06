package astinfo

import (
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
	content.WriteString("import (\"" + project.Module + "/gen\")\n")
	content.WriteString(`
func main() {
	wg:=gen.Run(gen.Config{
		Cors: true,
		Addr: ":8080",
		ServerName: "servlet",
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
