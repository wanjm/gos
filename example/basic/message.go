package basic

type ServletCode int
type Error struct {
	Code    ServletCode "json:\"code\""
	Message string      "json:\"message\""
}

func (error *Error) Error() string {
	return error.Message
}
