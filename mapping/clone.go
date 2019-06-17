package mapping

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

func Clone(ptrDst interface{}, src interface{}) error {
	return NamedClone(ptrDst, src, defaultNamer)
}

func NamedClone(ptrDst interface{}, src interface{}, namer Namer) error {
	ptrDstVal := reflect.ValueOf(ptrDst)
	if ptrDstVal.Kind() != reflect.Ptr {
		return fmt.Errorf("ptrDst is %v instead of reflect.Ptr", ptrDstVal.Type())
	}

	if ptrDstVal.IsNil() {
		return errors.New("ptrDst is nil")
	}

	if ptrDstVal.CanSet() {
		return errors.New("cannot set value to ptrDst")
	}

	return assignValue(ptrDstVal.Elem(), reflect.ValueOf(src), namer)
}
