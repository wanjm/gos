package astinfo

import (
	"go/ast"
	"strings"

	"github.com/wanjm/gos/astbasic"
)

const TagPrefix = "@gos"

// const GolangRawType = "rawType"

type Comment interface {
	dealValuePair(key, value string)
	// HasComment() bool
}

// 注释支持的格式为 @plaso url=xxx ; creator ; filter
func parseComment(commentGroup *ast.CommentGroup, commentor Comment) {
	if commentGroup == nil {
		return
	}

	for _, comment := range commentGroup.List {
		text := strings.TrimLeft(comment.Text, "/ \t") // 去掉前面的空格和斜杠
		if strings.HasPrefix(text, TagPrefix) {
			parseValidComment(text[len(TagPrefix):], commentor)
		}
	}
}

// 解析有效的comments
func parseValidComment(validComment string, commentor Comment) {
	commands := astbasic.Fields(validComment) // 多个参数以;分割
	for _, command := range commands {
		command = strings.Trim(command, " \t")
		if len(command) == 0 {
			continue
		}
		valuePair := strings.Split(command, "=") // 参数名和参数值以=分割
		valuePair[0] = strings.Trim(valuePair[0], " \t")
		// if len(valuePair) == 2 {
		// 	//去除前后空格和引号
		// 	valuePair[1] = strings.Trim(valuePair[1], " \t")
		// }
		if len(valuePair) == 2 {
			commentor.dealValuePair(valuePair[0], valuePair[1])
		} else {
			commentor.dealValuePair(valuePair[0], "")
		}
	}
}
