package meta

var DefaultConfig = &ServiceConfig{AdminName: "admin", AdminSecret: "admin"}
var Config ServiceConfig = *DefaultConfig

type ServiceConfig struct {
	AdminName   string
	AdminSecret string
}
