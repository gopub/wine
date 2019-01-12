package wine

type Config struct {
	Handlers []Handler
}

func DefaultConfig() *Config {
	c := &Config{}
	c.Handlers = []Handler{Logger()}
	return c
}
