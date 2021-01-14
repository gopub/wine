package wine

import (
	"github.com/gopub/conv"
	"github.com/gopub/log"
	"github.com/gopub/wine/internal/respond"
	"github.com/gopub/wine/router"
)

var logger *log.Logger

func init() {
	logger = log.Default().Derive("Wine")
	logger.SetFlags(log.LstdFlags - log.Lfunction - log.Lshortfile)
	router.SetLogger(logger)
	respond.SetLogger(logger)
}

func Logger() *log.Logger {
	return logger
}

type LogStringer interface {
	LogString() string
}

type Validator = conv.Validator

func Validate(i interface{}) error {
	return conv.Validate(i)
}

const (
	ParamNameDeviceID   = "device_id"
	ParamNameCoordinate = "coordinate"
	ParamNameTraceID    = "trace_id"
	ParamNameTimestamp  = "timestamp"
	ParamNameSign       = "sign"
	ParamNameAppID      = "app_id"
)
