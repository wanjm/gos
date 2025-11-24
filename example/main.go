package main
import ("flag"
"github.com/wan_jm/servlet_example/gen")

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
	/*
	下面的方法以来 github.com/wanjm/common 包，需要手动添加依赖；使用了优雅退出机制；
	common.InitLogger()
	manager := common.GracefulManager
	shutdown := gen.Start(gen.Config{
		Cors:       true,
		Addr:       ":8080",
		ServerName: "servlet", // this is the name of group tag in comments;
	})
	manager.Go("http server shutdown monitor", func(ctx context.Context) {
		shutdown(ctx, 5*time.Second)
	})
	manager.Wait()
 	*/
}
	