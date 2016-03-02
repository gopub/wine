package wine

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/jsix/gof/log"
	"github.com/justintan/xtypes"
	"io/ioutil"
	"net/http"
	"strings"
)

func parseHTTPReq(req *http.Request) (params xtypes.M) {
	params = xtypes.M{}
	for _, cookie := range req.Cookies() {
		params[cookie.Name] = cookie.Value
	}

	params.AddM(convertToM(req.URL.Query()))

	contentType := req.Header.Get("Content-Type")
	for i, ch := range contentType {
		if ch == ' ' || ch == ';' {
			contentType = contentType[:i]
			break
		}
	}

	switch contentType {
	case xtypes.MIMEHTML, xtypes.MIMEPlain:
		break
	case xtypes.MIMEJSON:
		d, e := ioutil.ReadAll(req.Body)
		if e != nil {
			log.PrintErr(e)
			break
		}

		if len(d) > 0 {
			var m xtypes.M
			e = jsonUnmarshal(d, &m)
			if e != nil {
				break
			}
			params.AddM(m)
		}
	case xtypes.MIMEPOSTForm:
		req.ParseForm()
		params.AddM(convertToM(req.Form))
	case xtypes.MIMEMultipartPOSTForm:
		req.ParseMultipartForm(32 << 20)
		if req.MultipartForm != nil && req.MultipartForm.File != nil {
			params.AddM(convertToM(req.MultipartForm.Value))
		}
	default:
		if len(contentType) > 0 {
			log.PrintErr(errors.New("unsupported content type"))
		}
		break
	}

	return
}

func convertToM(values map[string][]string) xtypes.M {
	m := xtypes.M{}
	for k, v := range values {
		i := strings.Index(k, "[]")
		if i >= 0 && i == len(k)-2 {
			k = k[0 : len(k)-2]
		}
		k = strings.ToLower(k)
		if len(v) > 1 || i >= 0 {
			m[k] = v
		} else if len(v) == 1 {
			m[k] = v[0]
		}
	}

	return m
}

func jsonUnmarshal(data []byte, pJSONObj interface{}) error {
	if len(data) == 0 {
		return errors.New("data is empty")
	}
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.UseNumber()
	err := decoder.Decode(pJSONObj)
	if err != nil {
		log.PrintErr(err)
	}
	return err
}
