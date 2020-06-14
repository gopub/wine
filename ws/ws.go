package ws

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/gopub/errors"
	"github.com/gopub/log"
	"github.com/gopub/wine"
)

var logger = wine.Logger()

func SetLogger(l *log.Logger) {
	logger = l
}

type GetAuthUserID interface {
	GetAuthUserID() int64
}

const (
	methodGetDate = "ws.getDate"
)

func MarshalData(v interface{}) (*Data, error) {
	if v == nil {
		return nil, nil
	}
	if d, ok := v.(*Data); ok {
		return d, nil
	}
	d := new(Data)
	if m, ok := v.(proto.Message); ok {
		b, err := proto.Marshal(m)
		if err != nil {
			return nil, err
		}
		d.V = &Data_Protobuf{
			Protobuf: b,
		}
		return d, nil
	} else {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		d.V = &Data_Json{
			Json: b,
		}
		return d, nil
	}
}

func (m *Data) Unmarshal(v interface{}) error {
	switch dv := m.V.(type) {
	case *Data_Json:
		return json.Unmarshal(dv.Json, v)
	case *Data_Protobuf:
		if m, ok := v.(proto.Message); ok {
			return proto.Unmarshal(dv.Protobuf, m)
		}
		return fmt.Errorf("v is %T not of proto.Message type", v)
	default:
		return fmt.Errorf("cannot unmarshal data type: %T", v)
	}
}

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

func NewReply(id int32, resultOrErr interface{}) *Reply {
	r := new(Reply)
	r.Id = id
	if err, ok := resultOrErr.(error); ok {
		code := errors.GetCode(err)
		if code <= 0 {
			code = http.StatusInternalServerError
		}
		r.Result = &Reply_Error{
			Error: &Error{
				Code:    int32(code),
				Message: err.Error(),
			},
		}
	} else {
		data, err := MarshalData(resultOrErr)
		if err != nil {
			r.Result = &Reply_Error{
				Error: &Error{
					Code:    int32(http.StatusInternalServerError),
					Message: err.Error(),
				},
			}
		} else {
			r.Result = &Reply_Data{
				Data: data,
			}
		}
	}
	return r
}

type PacketReadWriter interface {
	Read() (*Packet, error)
	Write(p *Packet) error
}
