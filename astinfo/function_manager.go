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
	}
}

type MethodManager struct {
	Server []*Method
}

func (m *MethodManager) AddCallable(callable Callable) {
	switch callable.GetType() {
	case Servlet:
		m.Server = append(m.Server, callable.(*Method))
	}
}
