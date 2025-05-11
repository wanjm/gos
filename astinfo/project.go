package astinfo

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Project 表示一个完整的Go项目结构
type Project struct {
	Name     string              // 项目名称（通过go.mod解析）
	Module   string              // 项目module名称
	Path     string              // 项目根目录的绝对路径
	Packages map[string]*Package // 项目包含的包集合
}

func NewProject() *Project {
	return &Project{
		Packages: make(map[string]*Package),
	}
}

// ParseMode 解析go.mod文件获取module名称
func (p *Project) ParseMode() error {
	modPath := filepath.Join(p.Path, "go.mod")
	modData, err := os.ReadFile(modPath)
	if err != nil {
		return fmt.Errorf("go.mod文件不存在: %v", err)
	}

	modLine := strings.Split(string(modData), "\n")[0]
	if !strings.HasPrefix(modLine, "module ") {
		return fmt.Errorf("无效的go.mod格式")
	}

	p.Module = strings.TrimSpace(strings.TrimPrefix(modLine, "module "))
	return nil
}

// ParsePackage 解析指定目录下的Go包
func (p *Project) ParsePackage(dirPath string) error {
	relPath, err := filepath.Rel(p.Path, dirPath)
	if err != nil {
		return err
	}

	pkg := NewPackage()
	pkg.Module = filepath.Join(p.Module, relPath)

	if err := pkg.Parse(dirPath); err != nil {
		return fmt.Errorf("解析包[%s]失败: %v", dirPath, err)
	}

	p.Packages[pkg.Module] = pkg
	return nil
}

// Parse 解析整个项目结构
func (p *Project) Parse(rootPath string) error {
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return err
	}
	p.Path = absPath

	if err := p.ParseMode(); err != nil {
		return err
	}

	return filepath.WalkDir(p.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}

		// 跳过隐藏目录和vendor目录
		if strings.HasPrefix(d.Name(), ".") || d.Name() == "vendor" {
			return fs.SkipDir
		}

		return p.ParsePackage(path)
	})
}
