package ws

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gopub/errors"
	"github.com/gorilla/websocket"
)

type Call struct {
	ID   int
	Name string
	Data []byte
}

func toCall(pc *Packet_Call) *Call {
	return &Call{
		ID:   int(pc.Id),
		Name: pc.Name,
		Data: pc.Data,
	}
}

type Reply struct {
	ID    int
	Data  []byte
	Error *errors.Error
}

func toReply(pr *Packet_Reply) *Reply {
	r := new(Reply)
	r.ID = int(pr.Id)
	switch v := pr.Result.(type) {
	case *Packet_Reply_Data:
		r.Data = v.Data
	case *Packet_Reply_Error:
		r.Error = errors.Format(int(v.Error.Code), v.Error.Message)
	}
	return r
}

type Conn struct {
	connMu sync.RWMutex
	conn   *websocket.Conn
	closed bool

	packetID    int32
	callC       chan *Call
	pushC       chan []byte
	readTimeout time.Duration

	replyChanM *sync.Map // map[int]chan<- *Reply

	userID int64
}

func NewConn(conn *websocket.Conn) *Conn {
	c := &Conn{
		conn:        conn,
		callC:       make(chan *Call, 256),
		pushC:       make(chan []byte, 256),
		replyChanM:  new(sync.Map),
		readTimeout: 10 * time.Second,
	}
	go c.startReadLoop()
	return c
}

func (c *Conn) Call(name string, params interface{}) (<-chan *Reply, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal params: %w", err)
	}
	call := &Packet_Call{
		Id:   atomic.AddInt32(&c.packetID, 1),
		Name: name,
		Data: data,
	}
	p := new(Packet)
	p.ContentType = Packet_JSON
	p.Content = &Packet_Call_{
		Call: call,
	}
	if err = c.write(p); err != nil {
		return nil, fmt.Errorf("cannot write packet: %w", err)
	}
	replyC := make(chan *Reply, 1)
	c.replyChanM.Store(int(call.Id), replyC)
	return replyC, nil
}

func (c *Conn) Push(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("cannot marshal json: %w", err)
	}
	p := new(Packet)
	p.ContentType = Packet_JSON
	p.Content = &Packet_Push{
		Push: data,
	}
	return c.write(p)
}

func (c *Conn) ReplyData(id int, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("cannot marshal json: %w", err)
	}
	r := &Packet_Reply{
		Id: int32(id),
		Result: &Packet_Reply_Data{
			Data: data,
		},
	}
	p := new(Packet)
	p.ContentType = Packet_JSON
	p.Content = &Packet_Reply_{
		Reply: r,
	}
	return c.write(p)
}

func (c *Conn) ReplyError(id int, err error) error {
	r := &Packet_Reply{
		Id: int32(id),
		Result: &Packet_Reply_Error{
			Error: &Packet_Error{
				Code:    int32(errors.GetCode(err)),
				Message: err.Error(),
			},
		},
	}
	p := new(Packet)
	p.ContentType = Packet_JSON
	p.Content = &Packet_Reply_{
		Reply: r,
	}
	return c.write(p)
}

func (c *Conn) write(p *Packet) error {
	data, err := proto.Marshal(p)
	if err != nil {
		return fmt.Errorf("cannot marshal packet: %v", err)
	}

	if c.closed {
		return errors.New("cannot write to closed conn")
	}

	c.connMu.Lock()
	err = c.conn.WriteMessage(websocket.BinaryMessage, data)
	c.connMu.Unlock()
	if err != nil {
		c.Close()
		return fmt.Errorf("cannot write binary message: %w", err)
	}
	return nil
}

func (c *Conn) CallC() <-chan *Call {
	return c.callC
}

func (c *Conn) startReadLoop() {
	for !c.closed {
		if err := c.conn.SetReadDeadline(time.Now().Add(c.readTimeout)); err != nil {
			logger.Errorf("Cannot set read deadline: %v", err)
			break
		}

		p := new(Packet)
		typ, data, err := c.conn.ReadMessage()
		if err != nil {
			logger.Errorf("Cannot read message: %v", err)
			break
		}

		if typ != websocket.BinaryMessage {
			logger.Errorf("Invalid message type: %d", typ)
			break
		}

		if err = proto.Unmarshal(data, p); err != nil {
			logger.Errorf("Cannot unmarshal packet: %v", err)
			break
		}

		switch v := p.Content.(type) {
		case *Packet_Call_:
			select {
			case c.callC <- toCall(v.Call):
				break
			default:
				logger.Errorf("Cannot write into call channel")
			}
		case *Packet_Reply_:
			id := int(v.Reply.Id)
			rc, ok := c.replyChanM.Load(id)
			if ok {
				c.replyChanM.Delete(id)
				rc.(chan<- *Reply) <- toReply(v.Reply)
			}
		case *Packet_Push:
			c.pushC <- v.Push
		}
	}
	c.Close()
}

func (c *Conn) Close() {
	if c.closed {
		return
	}
	c.connMu.Lock()
	if !c.closed {
		c.closed = true
		c.conn.Close()
	}
	c.connMu.Unlock()
}
