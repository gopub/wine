package wine

import (
	"fmt"
	"github.com/justintan/gox"
	"time"
)

func Logger(c Context) {
	st := time.Now()
	c.Next()
	cost := float32((time.Since(st) / time.Microsecond)) / 1000.0
	gox.LInfo(fmt.Sprintf("%.3fms %v%s%v %s",
		cost,
		gox.GreenColor,
		c.HttpRequest().Method,
		gox.ResetColor,
		c.HttpRequest().RequestURI))
}
