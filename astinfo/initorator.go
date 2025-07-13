package astinfo

import (
	"fmt"
	"strings"
)

const (
	globalPrefix = "__global_"
)

// 初始化函数依赖关系节点
type DependNode struct {
	Func               *Function
	Parent             []*DependNode
	returnVariableName string
}

// getReturnName 获取返回值名称
func (d *DependNode) getReturnName() string {
	return d.getReturnField().Name
}

// getReturnField 获取返回值字段
func (d *DependNode) getReturnField() *Field {
	return d.Func.Results[0]
}

type InitGroup struct {
	Initorators []*DependNode
	Default     *DependNode
}

// addNode 添加节点
func (g *InitGroup) addNode(node *DependNode) {
	g.Initorators = append(g.Initorators, node)
	if g.Default == nil {
		g.Default = node
	} else if node.getReturnName() == "" {
		if g.Default.getReturnName() == "" {
			fmt.Printf("more than one function return the same type,but without name %s %s\n", g.Default.Func.Name, node.Func.Name)
		} else {
			g.Default = node
		}
	}
}

type InitManager struct {
	variableMap VariableMap
	readyNode   []*DependNode
	project     *Project
}

// Generate(goGenerated *GenedFile) error
func (im *InitManager) Generate(goGenerated *GenedFile) error {
	if len(im.readyNode) == 0 {
		return nil
	}
	var definition strings.Builder
	var call strings.Builder
	definition.WriteString("var (\n")
	call.WriteString("func initVariable() {\n")
	for _, node := range im.readyNode {
		definition.WriteString(fmt.Sprintf("%s %s\n", node.returnVariableName, node.getReturnField().Type.Name()))
		call.WriteString(fmt.Sprintf("%s = %s\n", node.returnVariableName, node.Func.GenerateCallCode(goGenerated)))
	}
	definition.WriteString(")\n")
	call.WriteString("}\n")
	goGenerated.addBuilder(&definition)
	goGenerated.addBuilder(&call)
	im.project.initFuncs = append(im.project.initFuncs, "initVariable")
	return nil
}

func (p *Project) InitInitorator() {
	p.InitManager = &InitManager{
		variableMap: make(map[Typer]*InitGroup),
		project:     p,
	}
	p.InitManager.initInitorator()
}

// 返回初始化函数和map，key为Typer，value为相同返回值的数组
func (im *InitManager) collect() ([]*DependNode, VariableMap) {
	p := im.project
	functions := []*DependNode{}
	var waittingVariableMap VariableMap = make(map[Typer]*InitGroup)
	// 收集initiator到functions中；
	// 建立候选变量map
	for _, pkg := range p.Packages {
		for _, function := range pkg.Initiator {
			node := waittingVariableMap.addFunction(function)
			functions = append(functions, node)
		}
	}
	return functions, waittingVariableMap
}

// initInitorator 初始化初始化函数
func (im *InitManager) initInitorator() {
	// 创建variableMap
	functions, waittingVariableMap := im.collect()
	//将所有节点连接到父节点
	for _, node := range functions {
		im.initParent(node, waittingVariableMap)
	}
	// 每轮从functions中取出已经准备好了的function，放到ready的function中；
	var found bool = true
	for found {
		found = false
		var index int = 0
		for _, node := range functions {
			if im.variableMap.checkReady(node) {
				node.returnVariableName = globalPrefix + node.getReturnName()
				found = true
				im.variableMap.addNode(node)
				im.readyNode = append(im.readyNode, node)
			} else {
				functions[index] = node
				index++
			}
		}
		functions = functions[:index]
	}
	if len(functions) > 0 {
		for _, node := range functions {
			fmt.Printf("can't init function: %s\n", node.Func.Name)
		}
	}
}

// initParent 初始化父节点
func (im *InitManager) initParent(node *DependNode, waittingVariableMap VariableMap) {
	params := node.Func.Params
	for _, param := range params {
		parent := waittingVariableMap.getVariable(param.Type, param.Name)
		if parent != nil {
			node.Parent = append(node.Parent, parent)
		} else {
			fmt.Printf("can't init field: %s not found for function %s\n", param.Name, node.Func.Name)
		}
	}
}

type VariableMap map[Typer]*InitGroup

func (vm VariableMap) getVariable(typer Typer, name string) *DependNode {
	group := vm[typer]
	if group == nil {
		return nil
	}
	for _, initorator := range group.Initorators {
		if initorator.getReturnName() == name {
			return initorator
		}
	}
	return group.Default
}

// addFunction 添加初始化函数
func (im VariableMap) addFunction(function *Function) *DependNode {
	var node = &DependNode{
		Func: function,
	}
	im.addNode(node)
	return node
}

// addNode 添加节点
func (im VariableMap) addNode(node *DependNode) {
	typer := node.getReturnField().Type
	group := im[typer]
	if group == nil {
		group = &InitGroup{
			Initorators: []*DependNode{},
		}
		im[typer] = group
	}
	group.addNode(node)
}

// checkReady 检查是否所有父节点都已初始化
func (im VariableMap) checkReady(node *DependNode) bool {
	for _, parent := range node.Parent {
		if parent.returnVariableName == "" {
			return false
		}
	}
	return true
}
