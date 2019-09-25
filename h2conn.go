// +build ignore

package wine

import (
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/gopub/gox"
)

const h2connExpirationTime = time.Minute * 20
const h2connInitSize = 1024

type h2connEntry struct {
	id       string
	accessAt time.Time
}

type h2connCache struct {
	mu        sync.RWMutex
	conns     map[interface{}]*h2connEntry
	purgedAt  time.Time
	threshold int
}

func newH2ConnCache() *h2connCache {
	c := &h2connCache{}
	c.conns = make(map[interface{}]*h2connEntry, h2connInitSize)
	c.threshold = h2connInitSize
	return c
}

func (c *h2connCache) GetConnID(rw http.ResponseWriter) string {
	conn := GetHTTP2Conn(rw)
	if conn == nil {
		return ""
	}

	// Hope RLock is faster than Lock?
	c.mu.RLock()
	entry, ok := c.conns[conn]
	c.mu.RUnlock()
	if !ok {
		c.mu.Lock()
		entry, ok = c.conns[conn]
		if !ok {
			connID := gox.UniqueID()
			entry = &h2connEntry{
				id:       connID,
				accessAt: time.Now(),
			}
			c.conns[conn] = entry
			logger.Debugf("Detected new http/2 conn: %s, %v", connID, conn)

			if len(c.conns) > c.threshold || time.Since(c.purgedAt) > h2connExpirationTime {
				for k, v := range c.conns {
					if time.Since(v.accessAt) > h2connExpirationTime {
						delete(c.conns, k)
					}
				}

				if len(c.conns) > int(float64(c.threshold)*0.8) {
					c.threshold += h2connInitSize
				}

				c.purgedAt = time.Now()
			}
		}
		c.mu.Unlock()
	}

	entry.accessAt = time.Now()
	return entry.id
}

func GetHTTP2Conn(w http.ResponseWriter) interface{} {
	if reflect.TypeOf(w).String() != "*http.http2responseWriter" {
		return nil
	}
	http2responseWriter := reflect.ValueOf(w).Elem()
	http2responseWriterState := http2responseWriter.FieldByName("rws").Elem()
	conn := http2responseWriterState.FieldByName("conn").Elem().FieldByName("conn")
	return conn
}
