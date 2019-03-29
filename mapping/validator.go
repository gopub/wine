package mapping

import (
	"fmt"
	"github.com/gopub/gox"
	"reflect"
)

func Validate(model interface{}) gox.Error {
	val := reflect.ValueOf(model)
	if val.IsValid() == false {
		panic("not valid")
	}

	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	info := getModelInfo(val.Type())

	//for i, pi := range info {
	//	fmt.Printf("%d min=%v max=%v pattern=%s transformer=%s optional=%v\n", i, pi.minVal, pi.maxVal, pi.patterns, pi.transformName, pi.optional)
	//}

	for i := 0; i < val.NumField(); i++ {
		fv := val.Field(i)
		ft := val.Type().Field(i)
		if ft.Anonymous {
			err := Validate(fv.Interface())
			if err != nil {
				return err
			}
			continue
		}

		pi, ok := info[i]
		if !ok {
			continue
		}

		for fv.Kind() == reflect.Ptr && !fv.IsNil() {
			fv = fv.Elem()
		}

		if fv.Kind() == reflect.Ptr {
			if pi.optional {
				continue
			}

			return gox.BadRequest("missing parameter:" + pi.name)
		}

		switch fv.Kind() {
		case reflect.Slice:
			if fv.Len() == 0 {
				if pi.optional {
					continue
				}

				return gox.BadRequest("missing parameter:" + pi.name)
			}

			for i := 0; i < fv.Len(); i++ {
				err := Validate(fv.Index(i).Interface())
				if err != nil {
					return err
				}
			}
			continue
		case reflect.Map:
			if fv.Len() == 0 {
				if pi.optional {
					continue
				} else {
					return gox.BadRequest("missing parameter:" + pi.name)
				}
			}

			for _, k := range fv.MapKeys() {
				err := Validate(fv.MapIndex(k).Interface())
				if err != nil {
					return err
				}
			}
			continue
		case reflect.Struct:
			err := Validate(fv.Interface())
			if err != nil {
				return err
			}
			continue
		case reflect.String:
			if !pi.optional && len(fv.String()) == 0 {
				return gox.BadRequest("missing parameter:" + pi.name)
			}
		}

		//slice, map, struct don't support '=='
		if fv.Interface() == reflect.Zero(fv.Type()).Interface() && pi.optional {
			continue
		}

		if pi.minVal != nil {
			switch fv.Kind() {
			case reflect.Float32, reflect.Float64:
				if fv.Float() < pi.minVal.(float64) {
					return gox.BadRequest(fmt.Sprintf("%s's value must be larger than %v", pi.name, pi.minVal))
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if fv.Int() < pi.minVal.(int64) {
					return gox.BadRequest(fmt.Sprintf("%s's value must be larger than %v", pi.name, pi.minVal))

				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if fv.Uint() < pi.minVal.(uint64) {
					return gox.BadRequest(fmt.Sprintf("%s's value must be larger than %v", pi.name, pi.minVal))
				}
			case reflect.String:
				i := pi.minVal.(int64)
				if len(fv.String()) < int(i) {
					return gox.BadRequest(fmt.Sprintf("%s's len must be larger than %v", pi.name, pi.minVal))

				}
			default:
				panic("invalid kind: " + fv.Kind().String())
			}
		}

		if pi.maxVal != nil {
			switch fv.Kind() {
			case reflect.Float32, reflect.Float64:
				if fv.Float() > pi.maxVal.(float64) {
					return gox.BadRequest(fmt.Sprintf("%s's value must be less than %v", pi.name, pi.maxVal))
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if fv.Int() > pi.maxVal.(int64) {
					return gox.BadRequest(fmt.Sprintf("%s's value must be less than %v", pi.name, pi.maxVal))
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if fv.Uint() > pi.maxVal.(uint64) {
					return gox.BadRequest(fmt.Sprintf("%s's value must be less than %v", pi.name, pi.maxVal))
				}
			case reflect.String:
				i := pi.maxVal.(int64)
				if len(fv.String()) > int(i) {
					return gox.BadRequest(fmt.Sprintf("%s's len must be less than %v", pi.name, pi.maxVal))
				}
			default:
				panic("invalid kind: " + fv.Kind().String())
			}
		}

		for _, pattern := range pi.patterns {
			if !MatchPattern(pattern, fv.Interface()) {
				return gox.BadRequest(fmt.Sprintf("%s mismatches pattern=%s", pi.name, pattern))
			}
		}
	}
	return nil
}
