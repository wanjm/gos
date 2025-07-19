package astinfo

type Generation struct {
	ResponseKey string
	ResponseMod string
}
type Config struct {
	InitMain   bool
	Generation Generation
}
