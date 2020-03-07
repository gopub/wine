package stream

import (
	"context"
	"fmt"
	"github.com/gopub/log"
	"github.com/gopub/wine/internal/debug"
	"time"
)

func init() {
	debug.ByteStreamHandler = NewByteHandler(debugByteStream)
	debug.TextStreamHandler = NewTextHandler(debugTextStream)
	debug.JSONStreamHandler = NewJSONHandler(debugJSONStream)
}

func debugByteStream(ctx context.Context, w ByteWriteCloser) {
	i := 1
	for range time.Tick(time.Second) {
		v := fmt.Sprintf("%d.\t%v", i, time.Now())
		i++
		err := w.Write([]byte(v))
		if err != nil {
			w.Close()
			log.FromContext(ctx).Debug("Closed")
			break
		}
	}
}

func debugTextStream(ctx context.Context, w TextWriteCloser) {
	i := 1
	for range time.Tick(time.Second) {
		v := fmt.Sprintf("%d.\t%v", i, time.Now())
		i++
		err := w.Write(v)
		if err != nil {
			w.Close()
			log.FromContext(ctx).Debug("Closed")
			break
		}
	}
}

func debugJSONStream(ctx context.Context, w JSONWriteCloser) {
	i := 1
	for range time.Tick(time.Second) {
		v := struct {
			Seq  int       `json:"seq"`
			Time time.Time `json:"time"`
		}{i, time.Now()}
		i++
		err := w.Write(v)
		if err != nil {
			w.Close()
			log.FromContext(ctx).Debug("Closed")
			break
		}
	}
}
