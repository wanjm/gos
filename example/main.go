package main

import (
	"github.com/wan_jm/servlet_example/gen"
)

func main() {
	wg := gen.Run(gen.Config{
		Cors:       true,
		Addr:       ":8080",
		ServerName: "servlet",
	})
	wg.Wait()
}
