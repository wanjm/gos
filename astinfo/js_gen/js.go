package js_gen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wanjm/gos/astinfo"
	"github.com/wanjm/gos/basic"
)

type JsGen struct{}

func NewJsGen() *JsGen {
	return &JsGen{}
}

func (j *JsGen) GenerateCode(mp *astinfo.MainProject) {
	outDir := basic.Cfg.Generation.JsPath
	if outDir == "" {
		return
	}
	if !filepath.IsAbs(outDir) {
		outDir = filepath.Join(basic.Argument.SourcePath, outDir)
	}
	var err error
	outDir, err = filepath.Abs(outDir)
	if err != nil {
		fmt.Printf("Failed to resolve js gen directory: %v\n", err)
		return
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Printf("Failed to create js gen directory: %v\n", err)
		return
	}
	_ = mp
}
