package astinfo

import (
	"github.com/wanjm/gos/astbasic"
)

// 每个有自动生成代码的package 会有一个GenedFile类；
type GenedFile = astbasic.GenedFile

func CreateGenedFile(fileName string) *GenedFile {
	return GlobalProject.genPkg.NewFile(fileName)
}
