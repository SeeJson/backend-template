package httpserver

type Config struct {
	Address string `mapstructure:"address"`
}

var cfg Config

func SetConfig(c Config) {
	cfg = c
}

func Run() error {
	router := getRouter()
	return router.Run(cfg.Address)
}
