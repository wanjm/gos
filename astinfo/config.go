package astinfo

type Generation struct {
	ResponseKey string
	ResponseMod string
	AutoGen     bool
}
type Config struct {
	InitMain   bool
	Generation Generation
}
