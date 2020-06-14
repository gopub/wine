package ws

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gopub/errors"
	"github.com/gorilla/websocket"
)

func NewCall(id int32, name string, params interface{}) (*Call, error) {
	data, err := MarshalData(params)
	if err != nil {
		return nil, err
	}
	return &Call{
		Id:   id,
		Name: name,
		Data: data,
	}, nil
}

func NewDataReply(id int32, result interface{}) (*Reply, error) {
	data, err := MarshalData(result)
	if err != nil {
		return nil, err
	}
	return &Reply{
		Id: id,
		Result: &Reply_Data{
			Data: data,
		},
	}, nil
}

func NewErrorReply(id int32, err error) *Reply {
	return &Reply{
		Id: id,
		Result: &Reply_Error{
			Error: &Error{
				Code:    int32(errors.GetCode(err)),
				Message: err.Error(),
			},
		},
	}
}

type PacketReadWriter interface {
	Read() (*Packet, error)
	Write(p *Packet) error
}

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
		return nil, fmt.Errorf("cannot read message")
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
		return fmt.Errorf("cannot marshal packet: %w", err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	err = c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	if err != nil {
		return fmt.Errorf("cannot set write deadline: %w", err)
	}
	err = c.conn.WriteMessage(websocket.BinaryMessage, data)
	return errors.Wrapf(err, "cannot write binary message")
}

func (c *Conn) Call(ca *Call) error {
	return c.Write(&Packet{V: &Packet_Call{ca}})
}

func (c *Conn) WriteData(v interface{}) error {
	data, err := MarshalData(v)
	if err != nil {
		return fmt.Errorf("cannot marshal: %w", err)
	}
	return c.Write(&Packet{V: &Packet_Data{data}})
}

func (c *Conn) Reply(r *Reply) error {
	return c.Write(&Packet{V: &Packet_Reply{r}})
}

func (c *Conn) Close() {
	c.mu.Lock()
	c.conn.Close()
	c.mu.Unlock()
}
