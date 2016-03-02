package wine

import (
	"log"
	"time"
)

var (
	GreenColor   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	WhiteColor   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	YellowColor  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	RedColor     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	BlueColor    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	MagentaColor = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	CyanColor    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	ResetColor   = string([]byte{27, 91, 48, 109})
)

func Logger(c Context) {
	st := time.Now()
	c.Next()
	cost := float32((time.Since(st) / time.Microsecond)) / 1000.0
	log.Printf("%.3fms %v%s%v %s",
		cost,
		GreenColor,
		c.HTTPRequest().Method,
		ResetColor,
		c.HTTPRequest().RequestURI)
}
