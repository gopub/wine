package mapping

import (
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/gopub/gox"
)

const paramTag = "param"

type modelInfo = map[int]*paramInfo

var typeToModelInfo sync.Map

type paramInfo struct {
	name          string
	patterns      []string
	transformName string
	minVal        interface{}
	maxVal        interface{}
	optional      bool
}

func parseModelInfo(typ reflect.Type, tagName string) (modelInfo, error) {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		log.Panic("not struct")
	}

	indexToParamInfo := modelInfo{}

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		pi, err := parseParamInfo(f, tagName)
		if err != nil {
			panic(err.Error())
		}

		if pi == nil {
			continue
		}

		indexToParamInfo[i] = pi
	}

	return indexToParamInfo, nil
}

func parseParamInfo(f reflect.StructField, tagName string) (*paramInfo, error) {
	tag := strings.TrimSpace(strings.ToLower(f.Tag.Get(tagName)))
	if tag == "-" {
		return nil, nil
	}

	if f.Name[0] < 'A' || f.Name[0] > 'Z' {
		if len(tag) > 0 {
			return nil, errors.Errorf("unexported field: name=%s" + f.Name)
		}
		return nil, nil
	}

	info := &paramInfo{}
	if len(tag) == 0 {
		info.name = CamelToSnake(f.Name)
		return info, nil
	}

	strs := strings.Split(tag, ",")
	for _, s := range strs {
		s = strings.TrimSpace(s)
		if s == "optional" {
			info.optional = true
			continue
		}

		if MatchPattern("variable", s) {
			if len(info.name) > 0 {
				return nil, errors.Errorf("duplicate mapping name=%s", s)
			}
			info.name = s
			continue
		}

		kv := strings.SplitN(s, "=", 2)
		if len(kv) != 2 {
			return nil, errors.Errorf("invalid tag property=%s", s)
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		if len(key) == 0 || len(val) == 0 {
			return nil, errors.Errorf("invalid tag key=%s", key)
		}

		switch key {
		case "min":
			minVal, err := parseValueByType(f.Type, val)
			if err != nil {
				return nil, err
			}
			info.minVal = minVal
		case "max":
			maxVal, err := parseValueByType(f.Type, val)
			if err != nil {
				return nil, err
			}
			info.maxVal = maxVal
		case "pattern":
			if !MatchPattern("variable", val) {
				return nil, errors.Errorf("invalid pattern: name=%s", val)
			}

			info.patterns = append(info.patterns, val)
		case "trans", "transformer":
			if !MatchPattern("variable", val) {
				return nil, errors.Errorf("invalid transformer: name=%s", val)
			}

			if len(info.transformName) > 0 {
				return nil, errors.Errorf("duplicate transformer: name=%s", val)
			}

			info.transformName = val
		}
	}

	if len(info.name) == 0 {
		info.name = CamelToSnake(f.Name)
	}

	return info, nil
}

func parseValueByType(t reflect.Type, s string) (interface{}, error) {
	switch t.Kind() {
	case reflect.Float32, reflect.Float64:
		return gox.ParseFloat(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return gox.ParseInt(s)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return gox.ParseInt(s)
	case reflect.String:
		return gox.ParseInt(s)
	default:
		return nil, errors.New("min and max properties are not available for field type: " + t.String())
	}
}

func getModelInfo(modelType reflect.Type) modelInfo {
	if infoVal, ok := typeToModelInfo.Load(modelType); ok {
		return infoVal.(modelInfo)
	}

	info, err := parseModelInfo(modelType, paramTag)
	if err != nil {
		panic(err.Error())
	}

	typeToModelInfo.Store(modelType, info)
	return info
}
