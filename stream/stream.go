package stream

import (
	"context"
	"fmt"
	"github.com/gopub/wine"
	"time"

	"github.com/gopub/log"
)

const (
	Greeting = "WINE"
)

func InstallDebugRotes(r *wine.Router) {
	r.Get("bytestream", NewByteHandler(debugByteStream).(wine.HandlerFunc))
	r.Get("textstream", NewTextHandler(debugTextStream).(wine.HandlerFunc))
	r.Get("jsonstream", NewJSONHandler(debugJSONStream).(wine.HandlerFunc))
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
