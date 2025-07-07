package astinfo

// function & method 统称callable;
type Callable interface {
	GetType() string
}
type CallableManager interface {
	AddCallable(callable Callable)
}

type FunctionManager struct {
}

func (f *FunctionManager) AddCallable(callable Callable) {
}

type MethodManager struct {
}

func (m *MethodManager) AddCallable(callable Callable) {
}
