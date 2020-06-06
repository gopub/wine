package ws

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/gopub/types"

	"github.com/gorilla/websocket"
)

type ClientState int

const (
	Disconnected ClientState = iota
	Connecting
	Connected
)

type Client struct {
	connTimeout      time.Duration
	timeout          time.Duration
	pingInterval     time.Duration
	maxReconnBackoff time.Duration
	reconnBackoff    time.Duration

	addr string

	reqMu    sync.RWMutex
	reqs     *list.List
	reqChan  chan struct{}
	resChanM map[int64]chan<- *response

	connMu sync.Mutex
	conn   *websocket.Conn
	state  ClientState

	counter types.Counter
}

func NewClient(addr string) *Client {
	c := &Client{
		connTimeout:      10 * time.Second,
		timeout:          10 * time.Second,
		pingInterval:     10 * time.Second,
		maxReconnBackoff: 2 * time.Second,
		addr:             addr,
		reqs:             list.New(),
		resChanM:         make(map[int64]chan<- *response),
		state:            Disconnected,
	}
	go c.start()
	return c
}

func (c *Client) start() {
	for {
		c.run()
		if c.reconnBackoff > 0 {
			time.Sleep(c.reconnBackoff)
		}
		c.reconnBackoff += 100 * time.Millisecond
		if c.reconnBackoff > c.maxReconnBackoff {
			c.reconnBackoff = c.maxReconnBackoff
		}
	}
}

func (c *Client) run() {
	c.state = Connecting
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.addr, nil)
	if err != nil {
		cancel()
		logger.Errorf("Cannot dial %s: %v", c.addr, err)
		c.state = Disconnected
		return
	}
	cancel()
	c.conn = conn
	c.state = Connected
	done := make(chan struct{}, 1)
	go c.receiveLoop(done)
	c.sendLoop(done)
	c.conn.Close()
	c.state = Disconnected
}

func (c *Client) receiveLoop(done chan<- struct{}) {
	for {
		resp := new(response)
		err := c.conn.ReadJSON(resp)
		if err != nil {
			logger.Errorf("ReadJSON: %v", err)
			done <- struct{}{}
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

func (c *Client) sendLoop(done <-chan struct{}) {
	t := time.NewTicker(c.pingInterval)
	m := NewNetworkMonitor()
	defer m.Stop()
	defer t.Stop()
	for {
		select {
		case <-t.C:
			if err := c.conn.WriteJSON(&request{}); err != nil {
				logger.Errorf("Write: %v", err)
				c.reconnBackoff = 0
				return
			}
		case <-m.C:
			c.reconnBackoff = 0
			return
		case <-done:
			c.reconnBackoff = 0
			return
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

func (c *Client) SetConnTimeout(t time.Duration) {
	if t <= 0 {
		t = 0
	}
	c.connTimeout = t
}

func (c *Client) SetTimeout(t time.Duration) {
	if t <= time.Second {
		t = time.Second
	}
	c.timeout = t
}

func (c *Client) SetPingInterval(t time.Duration) {
	if t <= time.Second {
		t = time.Second
	}
	c.pingInterval = t
}

func (c *Client) SetMaxReconnBackoff(t time.Duration) {
	if t <= 0 {
		t = 0
	}
	c.maxReconnBackoff = t
}
