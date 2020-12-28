package websocket

import (
	"container/list"
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gopub/errors"
	"github.com/gorilla/websocket"
)

const ErrCanceled errors.String = "request canceled"

const (
	StatusTransportFailed = 600
	Status
)

type ClientState int

const (
	Disconnected ClientState = iota
	Connecting
	Connected
	Closed
)

func (s ClientState) String() string {
	switch s {
	case Disconnected:
		return "Disconnected"
	case Connecting:
		return "Connecting"
	case Connected:
		return "Connected"
	case Closed:
		return "Closed"
	default:
		return fmt.Sprint(int(s))
	}
}

type Caller interface {
	Call(ctx context.Context, name string, params interface{}, result interface{}) error
}

type Client struct {
	header           http.Header
	dialTimeout      time.Duration
	pingInterval     time.Duration
	maxReconnBackoff time.Duration
	reconnBackoff    time.Duration

	addr string

	newCallC chan struct{}
	mu       sync.RWMutex // guard calls, replyM
	calls    *list.List
	replyM   map[int32]chan<- *Reply
	metadata map[string]string

	conn    *Conn
	state   ClientState
	stateC  chan ClientState
	stateMu sync.Mutex

	callID int32

	Handshaker    func(rw PacketReadWriter) error
	Authenticator func(ctx context.Context, c Caller) error
	dataC         chan *Data
	pushC         chan *Push

	CallLogger func(call *Call, reply *Reply, callAt time.Time)
}

func NewClient(addr string, header http.Header) *Client {
	c := &Client{
		header:           header,
		dialTimeout:      10 * time.Second,
		pingInterval:     10 * time.Second,
		maxReconnBackoff: 2 * time.Second,
		addr:             addr,
		calls:            list.New(),
		newCallC:         make(chan struct{}, 256),
		replyM:           make(map[int32]chan<- *Reply),
		state:            Disconnected,
		stateC:           make(chan ClientState, 4),
		dataC:            make(chan *Data, 256),
		pushC:            make(chan *Push, 256),
		callID:           1,
		metadata:         map[string]string{},
	}
	c.CallLogger = c.logCall
	go c.start()
	return c
}

func (c *Client) nextCallID() int32 {
	atomic.AddInt32(&c.callID, 1)
	return c.callID
}

func (c *Client) start() {
	c.reconnBackoff = 100 * time.Millisecond
	for c.state != Closed {
		c.setState(Connecting)
		c.run()
		if c.state == Closed {
			break
		}
		c.setState(Disconnected)
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
	ctx, cancel := context.WithTimeout(context.Background(), c.dialTimeout)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.addr, c.header)
	if err != nil {
		cancel()
		logger.Errorf("Cannot connect %s: %v", c.addr, err)
		return
	}
	cancel()
	c.conn = NewConn(conn)
	if c.Handshaker != nil {
		if err = c.Handshaker(c.conn); err != nil {
			logger.Errorf("Cannot handshake: %v", err)
			return
		}
	}
	if err = c.writeMetadata(); err != nil {
		logger.Errorf("Cannot write metadata: %v", err)
		return
	}
	logger.Debugf("Wrote metadata: %v", c.metadata)
	c.setState(Connected)
	done := make(chan struct{}, 1)
	go c.read(done)
	go c.auth()
	c.write(done)
}

func (c *Client) read(done chan<- struct{}) {
	defer logger.Debug("Exited read loop")
	for {
		p, err := c.conn.Read()
		if err != nil {
			if c.state == Connected {
				logger.Errorf("Cannot read: %v", err)
			}
			done <- struct{}{}
			return
		}
		if v, ok := p.V.(*Packet_Data); ok {
			select {
			case c.dataC <- v.Data:
				break
			default:
				break
			}
		}
		switch v := p.V.(type) {
		case *Packet_Data:
			select {
			case c.dataC <- v.Data:
				break
			default:
				logger.Warnf("Data channel is overflow")
			}
		case *Packet_Push:
			select {
			case c.pushC <- v.Push:
				break
			default:
				logger.Warnf("Push channel is overflow")
			}
		case *Packet_Reply:
			c.mu.RLock()
			if ch, ok := c.replyM[v.Reply.Id]; ok {
				ch <- v.Reply
				delete(c.replyM, v.Reply.Id)
			}
			c.mu.RUnlock()
		}
	}
}

func (c *Client) write(done <-chan struct{}) {
	defer logger.Debug("Exited write loop")
	t := time.NewTicker(c.pingInterval)
	m := NewNetworkMonitor()
	defer m.Stop()
	defer t.Stop()
	for {
		select {
		case <-t.C:
			if err := c.conn.Hello(); err != nil {
				logger.Errorf("Cannot send hello: %v", err)
				c.reconnBackoff = 0
				return
			}
		case <-m.C:
			c.reconnBackoff = 0
			return
		case <-done:
			c.reconnBackoff = 0
			return
		case <-c.newCallC:
			c.mu.Lock()
		CallLoop:
			for it := c.calls.Front(); it != nil; {
				ca := it.Value.(*Call)
				next := it.Next()
				c.calls.Remove(it)
				it = next
				if err := c.conn.Call(ca.Id, ca.Name, ca.Data); err != nil {
					if c.state == Connected {
						logger.Errorf("Cannot call %s: %v", ca.Name, err)
					}
					if rc, ok := c.replyM[ca.Id]; ok {
						select {
						case rc <- NewReply(ca.Id, err):
							break
						default:
							break
						}
						delete(c.replyM, ca.Id)
					}
					break CallLoop
				}
			}
			c.mu.Unlock()
		}
	}
}

func (c *Client) writeMetadata() error {
	if len(c.metadata) == 0 {
		return nil
	}
	md := &Metadata{
		Entries: c.metadata,
	}
	p := new(Packet)
	p.V = &Packet_Metadata{
		Metadata: md,
	}
	return c.conn.Write(p)
}

func (c *Client) Call(ctx context.Context, name string, params interface{}, result interface{}) error {
	ca, err := NewCall(c.nextCallID(), name, params)
	if err != nil {
		return fmt.Errorf("cannot create call object: %w", err)
	}
	replyC := make(chan *Reply, 1)
	c.mu.Lock()
	c.calls.PushBack(ca)
	c.replyM[ca.Id] = replyC
	c.mu.Unlock()
	select {
	case c.newCallC <- struct{}{}:
		break
	default:
		return errors.New("too many pending calls")
	}

	startAt := time.Now()
	select {
	case <-ctx.Done():
		if c.CallLogger != nil {
			reply := new(Reply)
			reply.Id = ca.Id
			reply.Result = &Reply_Error{
				Error: &Error{
					Code:    StatusTransportFailed,
					Message: ctx.Err().Error(),
				},
			}
			c.CallLogger(ca, reply, startAt)
		}
		return ctx.Err()
	case reply := <-replyC:
		if c.CallLogger != nil {
			c.CallLogger(ca, reply, startAt)
		}
		switch v := reply.Result.(type) {
		case *Reply_Data:
			if result == nil {
				break
			}
			if err := v.Data.Unmarshal(result); err != nil {
				return fmt.Errorf("cannot unmarshal result: %w", err)
			}
		case *Reply_Error:
			if v.Error.Code == http.StatusUnauthorized {
				// Check flag in case recursive calling Authenticator
				if c.Authenticator != nil && ctx.Value(ckAuthFlag) == nil {
					// Reuse ctx, so total timeout equals to one call timeout
					ctx = context.WithValue(ctx, ckAuthFlag, true)
					err = c.Authenticator(ctx, c)
					if err == nil {
						logger.Debug("Authenticated")
						return c.Call(ctx, name, params, result)
					}
				}
			}
			return errors.Format(int(v.Error.Code), v.Error.Message)
		}
		return nil
	}
}

func (c *Client) CancelAll() {
	c.mu.Lock()
	for id, replyC := range c.replyM {
		select {
		case replyC <- NewReply(id, ErrCanceled):
			break
		default:
			break
		}
	}
	c.replyM = make(map[int32]chan<- *Reply)
	c.mu.Unlock()
}

func (c *Client) SetMetadata(h map[string]string) {
	c.mu.Lock()
	for k, v := range h {
		c.metadata[k] = v
	}
	c.mu.Unlock()
	if c.state == Connected {
		logger.Debugf("Writing metadata: %v", h)
		go c.writeMetadata()
	} else {
		logger.Debugf("Delay writing metadata: %v", h)
	}
}

func (c *Client) Close() {
	c.setState(Closed)
	close(c.dataC)
	close(c.stateC)
}

func (c *Client) setState(s ClientState) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	if c.state == s {
		return
	}
	if c.state == Closed {
		logger.Errorf("Cannot change state from %v to %v", c.state, s)
		return
	}
	c.state = s
	if s == Closed && c.conn != nil {
		c.conn.Close()
	}

	select {
	case c.stateC <- s:
		logger.Debugf("State: %v", s)
	default:
		logger.Warnf("State: %v, cannot write to overflow chan", s)
	}
}

func (c *Client) State() ClientState {
	return c.state
}

func (c *Client) StateC() <-chan ClientState {
	return c.stateC
}

func (c *Client) SetDialTimeout(t time.Duration) {
	if t < time.Second {
		t = time.Second
	}
	c.dialTimeout = t
}

func (c *Client) SetPingInterval(t time.Duration) {
	if t < time.Second {
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

func (c *Client) DataC() <-chan *Data {
	return c.dataC
}

func (c *Client) PushC() <-chan *Push {
	return c.pushC
}

func (c *Client) GetServerTime(ctx context.Context) (time.Time, error) {
	var res struct {
		Timestamp int64 `json:"timestamp"`
	}
	err := c.Call(ctx, methodGetDate, nil, &res)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(res.Timestamp, 0), nil
}

func (c *Client) auth() {
	if c.Authenticator == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.dialTimeout)
	defer cancel()
	ctx = context.WithValue(ctx, ckAuthFlag, true)
	err := c.Authenticator(ctx, c)
	if err == nil {
		logger.Debug("Authenticated")
	}
}

func (c *Client) logCall(call *Call, reply *Reply, callAt time.Time) {
	cost := time.Since(callAt)
	switch v := reply.Result.(type) {
	case *Reply_Data:
		logger.Debugf("#%d %s %v", call.Id, call.Name, cost)
	case *Reply_Error:
		logger.Errorf("#%d %s %v | %s | %d:%s", call.Id, call.Name, cost, call.Data.LogString(), v.Error.Code, v.Error.Message)
	}
}
