package astinfo

// function & method 统称callable;
type Callable interface {
	GetType() string
}
type CallableManager interface {
	AddCallable(callable Callable)
}

type FunctionManager struct {
	Initiator []*Function
	Filter    []*Function
}

func (f *FunctionManager) AddCallable(callable Callable) {
	switch callable.GetType() {
	case Initiator:
		f.Initiator = append(f.Initiator, callable.(*Function))
	case FilterConst:
		f.Filter = append(f.Filter, callable.(*Function))
	case Creator:
	case Websocket:
	}
}

type MethodManager struct {
	Server []*Method
}

func (m *MethodManager) AddCallable(callable Callable) {
	// 当前这个type是servlet，时parseFunction写死的。
	// 注意此处的servlet，filter，都是针对struct的type的子类型来的。
	// 所以有 raw::servlet; servlet::servlet;
	switch callable.GetType() {
	case Servlet:
		m.Server = append(m.Server, callable.(*Method))
	}
}
