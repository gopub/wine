package ws

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
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

func UnmarshalData(data *Data, v interface{}) error {
	if data == nil {
		if v == nil {
			return nil
		}
		return errors.New("data is nil")
	}
	switch dv := data.V.(type) {
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
