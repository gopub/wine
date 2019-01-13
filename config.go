package wine

type Config struct {
	Handlers []Handler
}

func DefaultConfig() *Config {
	c := &Config{}
	return c
}
