package websocket

import (
	"fmt"
	"sync"
	"time"

	"github.com/gopub/errors"
	"github.com/gopub/log"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type Conn struct {
	mu           sync.RWMutex
	conn         *websocket.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewConn(conn *websocket.Conn) *Conn {
	c := &Conn{
		conn:         conn,
		readTimeout:  20 * time.Second,
		writeTimeout: 10 * time.Second,
	}
	return c
}

func (c *Conn) Read() (*Packet, error) {
	if err := c.conn.SetReadDeadline(time.Now().Add(c.readTimeout)); err != nil {
		return nil, fmt.Errorf("cannot set read deadline: %w", err)
	}
	t, data, err := c.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("cannot read message: %w", err)
	}
	if t != websocket.BinaryMessage {
		return nil, fmt.Errorf("expect message type %d got %d", websocket.BinaryMessage, t)
	}
	p := new(Packet)
	if err = proto.Unmarshal(data, p); err != nil {
		return nil, fmt.Errorf("cannot unmarshal packet: %w", err)
	}
	return p, nil
}

func (c *Conn) Write(p *Packet) error {
	data, err := proto.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal packet: %w", err)
	}
	c.mu.Lock()
	if c.conn == nil {
		return errors.New("cannot write to a closed conn")
	}
	defer c.mu.Unlock()
	err = c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	if err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}
	err = c.conn.WriteMessage(websocket.BinaryMessage, data)
	return errors.Wrapf(err, "write binary message")
}

func (c *Conn) Call(id int32, name string, params interface{}) error {
	ca, err := NewCall(id, name, params)
	if err != nil {
		return err
	}
	return c.Write(&Packet{V: &Packet_Call{ca}})
}

func (c *Conn) Push(typ int32, data interface{}) error {
	p, err := NewPushPacket(typ, data)
	if err != nil {
		return err
	}
	return c.Write(p)
}

func (c *Conn) WriteData(v interface{}) error {
	p, err := NewDataPacket(v)
	if err != nil {
		return err
	}
	return c.Write(p)
}

func (c *Conn) Reply(id int32, resultOrErr interface{}) error {
	return c.Write(&Packet{V: &Packet_Reply{NewReply(id, resultOrErr)}})
}

func (c *Conn) Hello() error {
	return c.Write(&Packet{V: &Packet_Hello{Hello: new(Hello)}})
}

func (c *Conn) Close() {
	if c.conn == nil {
		log.Warn("Already closed")
		return
	}
	c.mu.Lock()
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Errorf("Close websocket conn: %w", err)
		}
		c.conn = nil
	}
	c.mu.Unlock()
}
