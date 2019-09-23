package wine

import (
	"github.com/gopub/gox"
	"github.com/gopub/log"
	"net/http"
	"time"
)

type Config struct {
	Handlers        []Handler
	ParamsParser    ParamsParser
	RequestTimeout  time.Duration
	ResponseHeaders http.Header
	Logger          log.Logger
}

func DefaultConfig() *Config {
	logger := log.GetLogger("Wine")
	logger.SetFlags(logger.Flags() ^ log.Lfunction ^ log.Lshortfile)
	header := make(http.Header, 1)
	header.Set("Server", "Wine")
	c := &Config{
		Logger:          logger,
		ParamsParser:    NewDefaultParamsParser([]string{"sid", "device_id"}, 8*gox.MB),
		RequestTimeout:  10 * time.Second,
		ResponseHeaders: header,
	}
	return c
}
