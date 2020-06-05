package ws

import (
	"container/list"
	"context"
	"github.com/gopub/types"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ConnState int

const (
	Disconnected ConnState = iota
	Connecting
	Connected
)

type Client struct {
	Timeout      time.Duration
	PingInterval time.Duration
	addr         string

	reqMu    sync.RWMutex
	reqs     *list.List
	reqChan  chan struct{}
	resChanM map[int64]chan<- *response

	connMu    sync.Mutex
	conn      *websocket.Conn
	connState ConnState

	counter types.Counter
}

func NewClient(addr string) *Client {
	c := &Client{
		Timeout:      10 * time.Second,
		PingInterval: 10 * time.Second,
		addr:         addr,
		reqs:         list.New(),
		resChanM:     make(map[int64]chan<- *response),
		connState:    Disconnected,
	}
	return c
}

func (c *Client) runLoop() {
	maxInterval := 10 * time.Second
	interval := time.Second
	for {
		c.run()
		time.Sleep(interval)
		interval += time.Second
		if interval > maxInterval {
			interval = maxInterval
		}
	}
}

func (c *Client) run() {
	c.setConnState(Connecting)
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.addr, nil)
	if err != nil {
		cancel()
		logger.Errorf("Cannot dial %s: %v", c.addr, err)
		c.setConnState(Disconnected)
		return
	}
	cancel()
	c.conn = conn
	c.setConnState(Connected)
	go c.receiveLoop()
	c.sendLoop()
}

func (c *Client) receiveLoop() {
	for {
		resp := new(response)
		err := c.conn.ReadJSON(resp)
		if err != nil {
			logger.Errorf("ReadJSON: %v", err)
			return
		}
		c.reqMu.RLock()
		if ch, ok := c.resChanM[resp.ID]; ok {
			ch <- resp
			delete(c.resChanM, resp.ID)
		}
		c.reqMu.RUnlock()
	}
}

func (c *Client) sendLoop() {
	defer c.Disconnect()
	t := time.NewTicker(c.PingInterval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			if err := c.conn.WriteJSON(&request{}); err != nil {
				logger.Errorf("Write: %v", err)
				return
			}
		case <-c.reqChan:
			c.reqMu.Lock()
			for it := c.reqs.Front(); it != nil; it = it.Next() {
				req := it.Value.(*request)
				if err := c.conn.WriteJSON(req); err != nil {
					logger.Errorf("Write %s: %v", req.Name, err)
					c.reqMu.Unlock()
					return
				}
			}
			c.reqMu.Unlock()
		}
	}
}

func (c *Client) setConnState(s ConnState) {
	if c.connState == s {
		return
	}
	c.connState = s
}

func (c *Client) Send(ctx context.Context, name string, data interface{}) (interface{}, error) {
	req := &request{
		ID:   c.counter.Next(),
		Name: name,
		Data: data,
	}
	resChan := make(chan *response, 1)
	defer close(resChan)
	c.reqMu.Lock()
	c.reqs.PushBack(req)
	c.resChanM[req.ID] = resChan
	c.reqMu.Unlock()
	select {
	case c.reqChan <- struct{}{}:
		break
	default:
		break
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-resChan:
		return res.Data, res.Error
	}
}

func (c *Client) Disconnect() {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.setConnState(Disconnected)
}
