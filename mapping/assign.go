package mapping

import (
	"reflect"

	"github.com/pkg/errors"

	"github.com/gopub/gox"
	"github.com/gopub/log"
)

type Assigner interface {
	Assign(v interface{}) error
}

type Validator interface {
	Validate() error
}

// Assign assigns src to dst
func Assign(dst interface{}, src interface{}) error {
	return AssignWithNamer(dst, src, defaultNamer)
}

// Assign assigns src to dst with namer
func AssignWithNamer(dst interface{}, src interface{}, namer Namer) error {
	if namer == nil {
		return errors.New("namer is nil")
	}

	dstVal := reflect.ValueOf(dst)
	if dstVal.IsValid() == false {
		return errors.New("dst is invalid")
	}

	for dstVal.Kind() == reflect.Ptr && !dstVal.IsNil() {
		dstVal = dstVal.Elem()
	}

	// dst must be a nil pointer or a valid value
	err := assignValue(dstVal, reflect.ValueOf(src), namer)
	return errors.Wrap(err, "cannot assignValue")
}

// assignValue dst is valid value or pointer to value
func assignValue(dst reflect.Value, src reflect.Value, namer Namer) error {
	if !src.IsValid() {
		return errors.New("src is invalid")
	}

	if !dst.IsValid() {
		return errors.Errorf("invalid values:dst=%#v,src=%#v", dst, src)
	}

	if a, ok := dst.Interface().(Assigner); ok {
		return a.Assign(src.Interface())
	}

	v := dst
	if v.Kind() == reflect.Ptr {
		if v.IsNil() && v.CanSet() {
			v = reflect.New(v.Type().Elem())
		}
		v = v.Elem()
	}

	for src.Kind() == reflect.Ptr || src.Kind() == reflect.Interface {
		if src.IsNil() {
			return nil
		}
		src = src.Elem()
	}

	if !v.CanSet() {
		return errors.New("cannot set")
	}

	switch v.Kind() {
	case reflect.Bool:
		b, err := gox.ParseBool(src.Interface())
		if err != nil {
			return errors.Wrap(err, "cannot ParseBool")
		}
		v.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := gox.ParseInt(src.Interface())
		if err != nil {
			return errors.Wrap(err, "cannot ParseInt")
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := gox.ParseInt(src.Interface())
		if err != nil {
			return errors.Wrap(err, "cannot ParseInt")
		}
		v.SetUint(uint64(i))
	case reflect.Float32, reflect.Float64:
		i, err := gox.ParseFloat(src.Interface())
		if err != nil {
			return errors.Wrap(err, "cannot ParseFloat")
		}
		v.SetFloat(i)
	case reflect.String:
		if src.Kind() != reflect.String {
			return errors.New("src isn't string")
		}
		v.SetString(src.String())
	case reflect.Slice:
		if src.Kind() != reflect.Slice {
			return errors.New("src isn't slice")
		}
		v.Set(reflect.MakeSlice(v.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			err := assignValue(v.Index(i), src.Index(i), namer)
			if err != nil {
				return errors.Wrapf(err, "cannot assignValue: i=%d", i)
			}
		}
	case reflect.Map:
		err := mapToMap(v, src, namer)
		if err != nil {
			return errors.Wrapf(err, "cannot mapToMap")
		}
	case reflect.Struct:
		if src.Kind() == reflect.Map {
			if err := mapToStruct(v, src, namer); err != nil {
				return errors.Wrapf(err, "cannot mapToStruct")
			}
		} else if src.Kind() == reflect.Struct {
			if err := structToStruct(v, src, namer); err != nil {
				return errors.Wrapf(err, "cannot structToStruct")
			}
		} else {
			return errors.Errorf("src is %v instead of struct or map", src.Kind())
		}
	default:
		return errors.Errorf("Unexpected dst %v", v.Kind())
	}

	if dst.Kind() == reflect.Ptr && dst.IsNil() {
		dst.Set(v.Addr())
	}

	if vr, ok := dst.Interface().(Validator); ok {
		return vr.Validate()
	}
	return nil
}

// both dst and src must be map
func mapToMap(dst reflect.Value, src reflect.Value, namer Namer) error {
	if dst.Kind() != reflect.Map {
		return errors.Errorf("dst isn't map: kind=%s", dst.Kind().String())
	}

	if src.Kind() != reflect.Map {
		return errors.New("src isn't map")
	}

	if !src.Type().Key().AssignableTo(dst.Type().Key()) {
		return errors.Errorf("src:key=%s,type=%s can't be assigned to dst:key=%s,type=%s",
			src.Type().Key().String(), src.Type().String(),
			dst.Type().Key().String(), src.Type().String())
	}

	if dst.IsNil() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}

	de := dst.Type().Elem()
	canAssign := src.Type().Elem().AssignableTo(de)
	for _, k := range src.MapKeys() {
		switch {
		case canAssign:
			dst.SetMapIndex(k, src.MapIndex(k))
		case de.Kind() == reflect.Ptr:
			kv := reflect.New(de.Elem())
			err := assignValue(kv, src.MapIndex(k), namer)
			if err != nil {
				log.Warnf("cannot assignValue: %v", err)
				continue
			}
			dst.SetMapIndex(k, kv)
		default:
			kv := reflect.New(de)
			err := assignValue(kv, src.MapIndex(k), namer)
			if err != nil {
				log.Warnf("cannot assignValue: %v", err)
				continue
			}
			dst.SetMapIndex(k, kv.Elem())
		}
	}
	return nil
}

func mapToStruct(dst reflect.Value, src reflect.Value, namer Namer) error {
	if dst.Kind() != reflect.Struct {
		return errors.Errorf("dst isn't struct: kind=%v", dst.Kind())
	}

	if src.Kind() == reflect.Interface {
		src = src.Elem()
	}

	if src.Kind() != reflect.Map {
		return errors.Errorf("src: kind=%v isn't map", src.Kind())
	}

	if src.Type().Key().Kind() != reflect.String {
		return errors.Errorf("key: type=%v must be string", src.Type().Key().Kind())
	}

	for i := 0; i < dst.NumField(); i++ {
		fieldVal := dst.Field(i)
		if fieldVal.IsValid() == false || fieldVal.CanSet() == false {
			continue
		}

		fieldType := dst.Type().Field(i)
		if fieldType.Anonymous {
			err := assignValue(fieldVal, src, namer)
			if err != nil {
				log.Warnf("cannot assignValue: i=%d %v", i, err)
			}
			continue
		}

		for _, key := range src.MapKeys() {
			if namer.Name(key.String()) != fieldType.Name {
				continue
			}

			fieldSrcVal := src.MapIndex(key)
			if !fieldSrcVal.IsValid() {
				log.Warnf("field: name=%s is invalid", fieldType.Name)
				continue
			}

			err := assignValue(fieldVal, reflect.ValueOf(fieldSrcVal.Interface()), namer)
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func structToStruct(dst reflect.Value, src reflect.Value, namer Namer) error {
	if dst.Kind() != reflect.Struct {
		log.Panicf("dst: kind=%s isn't struct", dst.Kind())
	}

	if src.Kind() == reflect.Interface {
		src = src.Elem()
	}

	if src.Kind() != reflect.Struct {
		return errors.New("src isn't struct")
	}

	for i := 0; i < dst.NumField(); i++ {
		dstFieldVal := dst.Field(i)
		if dstFieldVal.IsValid() == false || dstFieldVal.CanSet() == false {
			continue
		}

		dstFieldType := dst.Type().Field(i)
		if dstFieldType.Anonymous {
			err := assignValue(dstFieldVal, src, namer)
			if err != nil {
				log.Warnf("cannot assignValue: %v", err)
			}
			continue
		}

		for i := 0; i < src.NumField(); i++ {
			srcFieldVal := src.Field(i)
			srcFieldName := src.Type().Field(i).Name
			if srcFieldVal.IsValid() == false || srcFieldName[0] < 'A' || srcFieldName[0] > 'Z' {
				continue
			}

			if namer.Name(srcFieldName) != dstFieldType.Name {
				continue
			}

			err := assignValue(dstFieldVal, reflect.ValueOf(srcFieldVal.Interface()), namer)
			if err != nil {
				log.Warnf("cannot assignValue: %s %v", dstFieldType.Name, err)
			}
			break
		}
	}

	for i := 0; i < src.NumField(); i++ {
		sfv := src.Field(i)
		sfName := src.Type().Field(i).Name
		if sfv.IsValid() == false || sfName[0] < 'A' || sfName[0] > 'Z' {
			continue
		}

		if src.Type().Field(i).Anonymous {
			if err := assignValue(dst, reflect.ValueOf(sfv.Interface()), namer); err != nil {
				log.Warnf("cannot assignValue: %s %v", sfName, err)
			}
		}
	}
	return nil
}
