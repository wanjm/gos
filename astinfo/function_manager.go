package astinfo

type FunctionManag interface {
	addServlet(*Function)
	addCreator(childClass *Struct, method *Function)
	addInitiator(initiator *Function)
}

type FunctionManager struct {
	creators   map[*Struct]*Function //纪录构建默认参数的代码, key是构建的struct
	initiators []*DependNode         //初始化函数依赖关系
	servlets   []*Function           //记录路由代码
	postAction map[string]*Function  //记录后置操作
}

func createFunctionManager() FunctionManager {
	return FunctionManager{
		creators:   make(map[*Struct]*Function),
		postAction: make(map[string]*Function),
	}
}

func (funcManager *FunctionManager) addServlet(function *Function) {
	funcManager.servlets = append(funcManager.servlets, function)
}

func (funcManager *FunctionManager) addCreator(childClass *Struct, function *Function) {
	funcManager.creators[childClass] = function
}

// 入参直接是函数返回值的对象，跟method.Result[0]相同,为了保持返回值的variable不受影响
func (funcManager *FunctionManager) addInitiator(initiator *Function) {
	funcManager.initiators = append(
		funcManager.initiators,
		&DependNode{
			function: initiator,
		},
	)
}

func (funcManager *FunctionManager) getCreator(childClass *Struct) (function *Function) {
	return funcManager.creators[childClass]
}
