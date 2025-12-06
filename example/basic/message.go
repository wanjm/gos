package basic
type Error struct {
	Code    int    "json:\"code\""
	Message string "json:\"message\""
}

func (error *Error) Error() string {
	return error.Message
}
func New(code int, msg string) error {
	res := &Error{
		Code:    code,
		Message: msg,
	}
	return res
}
func (error *Error) GetErrorCode() int {
	return error.Code
}
	