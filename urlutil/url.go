package urlutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/gopub/conv"
)

func Join(a ...string) string {
	if len(a) == 0 {
		return ""
	}
	p := path.Join(a...)
	p = strings.Replace(p, ":/", "://", 1)
	i := strings.Index(p, "://")
	s := p
	if i >= 0 {
		s = p[i:]
		l := strings.Split(s, "/")
		for i, v := range l {
			l[i] = url.PathEscape(v)
		}
		p = p[:i] + path.Join(l...)
	}
	return p
}

func ToURL(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("parse: %w", err)
	}
	if u.Scheme == "" {
		return "", errors.New("missing schema")
	}
	if u.Host == "" {
		return "", errors.New("missing host")
	}
	return u.String(), nil
}

func IsURL(s string) bool {
	u, _ := ToURL(s)
	return u != ""
}

func VarargsToURLValues(keyAndValues ...interface{}) (url.Values, error) {
	uv := url.Values{}
	keys, vals, err := conv.VarargsToSlice(keyAndValues)
	if err != nil {
		return nil, err
	}
	for i, k := range keys {
		vs, err := conv.ToString(vals[i])
		if err != nil {
			return nil, err
		}
		if vs != "" {
			uv.Add(k, vs)
		}
	}
	return uv, nil
}

func MustVarargsToURLValues(keyAndValues ...interface{}) url.Values {
	v, err := VarargsToURLValues(keyAndValues...)
	if err != nil {
		panic(err)
	}
	return v
}

func ToValues(i interface{}) (url.Values, error) {
	i = conv.IndirectToStringerOrError(i)
	if i == nil {
		return nil, errors.New("nil values")
	}
	switch v := i.(type) {
	case url.Values:
		return v, nil
	}

	b, err := json.Marshal(i)
	if err != nil {
		return nil, fmt.Errorf("cannot convert %#v of type %T to url.Values", i, i)
	}
	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, fmt.Errorf("cannot convert %#v of type %T to url.Values", i, i)
	}
	uv := url.Values{}
	for k, v := range m {
		uv.Set(k, fmt.Sprint(v))
	}
	return uv, nil
}

func MustValues(i interface{}) url.Values {
	v, err := ToValues(i)
	if err != nil {
		panic(err)
	}
	return v
}
