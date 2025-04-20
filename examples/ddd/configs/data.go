package configs

type Config struct {
	Stage       string            `mapstructure:"stage"`
	Servers     map[string]Server `mapstructure:"servers"`
	GraylogAddr string            `mapstructure:"graylog_address"`
}

type Server struct {
	Address   string `mapstructure:"address,omitempty"`
	ProxyPath string `mapstructure:"proxy_path,omitempty"`
}
