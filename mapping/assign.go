package mapping

import (
	"reflect"
	"strings"

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

	err := assignValue(reflect.ValueOf(dst), reflect.ValueOf(src), namer)
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
		if dst.Kind() != reflect.Ptr || !dst.IsNil() {
			return a.Assign(src.Interface())
		}

		if dst.Kind() == reflect.Ptr && dst.CanSet() {
			dst.Set(reflect.New(dst.Type().Elem()))
			if err := dst.Interface().(Assigner).Assign(src.Interface()); err != nil {
				dst.Set(reflect.Zero(dst.Type()))
				return errors.Wrap(err, "cannot assign via Assigner interface")
			}
			return nil
		}
		return errors.New("cannot assign")
	}

	v := dst
	for v.Kind() == reflect.Ptr {
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

// mapToStruct assign map to struct, panic if src is not map or dst is not struct
func mapToStruct(dst reflect.Value, src reflect.Value, namer Namer) error {
	if dst.Kind() != reflect.Struct {
		log.Panicf("dst is %v instead of struct", dst.Kind())
	}

	if src.Kind() != reflect.Map {
		log.Panicf("src is %v instead of struct", src.Kind())
	}

	if src.Type().Key().Kind() != reflect.String {
		return errors.Errorf("key: type=%v must be string", src.Type().Key().Kind())
	}

	for i := 0; i < dst.NumField(); i++ {
		fieldVal := dst.Field(i)
		if !fieldVal.IsValid() || !fieldVal.CanSet() {
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

		lcFieldName := strings.ToLower(fieldType.Name)
		for _, key := range src.MapKeys() {
			if strings.ToLower(namer.Name(key.String())) != lcFieldName {
				continue
			}

			fieldSrcVal := src.MapIndex(key)
			if !fieldSrcVal.IsValid() {
				log.Warnf("field: name=%s is invalid", fieldType.Name)
				continue
			}

			err := assignValue(fieldVal, reflect.ValueOf(fieldSrcVal.Interface()), namer)
			if err != nil {
				return errors.Wrapf(err, "cannot assignValue: %s", key.String())
			}
			break
		}
	}
	return nil
}

// structToStruct assign struct to struct, panic if src or dst is not struct
func structToStruct(dst reflect.Value, src reflect.Value, namer Namer) error {
	if dst.Kind() != reflect.Struct {
		log.Panicf("dst is %v instead of struct", dst.Kind())
	}

	if src.Kind() != reflect.Struct {
		log.Panicf("src is %v instead of struct", dst.Kind())
	}

	for i := 0; i < dst.NumField(); i++ {
		dstFieldVal := dst.Field(i)
		if !dstFieldVal.IsValid() || !dstFieldVal.CanSet() {
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

		lcDstFieldName := strings.ToLower(dstFieldType.Name)
		for i := 0; i < src.NumField(); i++ {
			srcFieldVal := src.Field(i)
			srcFieldName := src.Type().Field(i).Name
			if !srcFieldVal.IsValid() || srcFieldName[0] < 'A' || srcFieldName[0] > 'Z' {
				continue
			}

			if strings.ToLower(namer.Name(srcFieldName)) != lcDstFieldName {
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
		srcFieldVal := src.Field(i)
		srcFieldName := src.Type().Field(i).Name
		if !srcFieldVal.IsValid() || srcFieldName[0] < 'A' || srcFieldName[0] > 'Z' {
			continue
		}

		if src.Type().Field(i).Anonymous {
			if err := assignValue(dst, reflect.ValueOf(srcFieldVal.Interface()), namer); err != nil {
				log.Warnf("cannot assignValue: %s %v", srcFieldName, err)
			}
		}
	}
	return nil
}
