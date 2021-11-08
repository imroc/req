/*
  GoLang code created by Jirawat Harnsiriwatanakit https://github.com/kazekim
*/

package req

import "encoding/json"

// Header represents http request header
type Header map[string]string

func (h Header) Clone() Header {
	if h == nil {
		return nil
	}
	hh := Header{}
	for k, v := range h {
		hh[k] = v
	}
	return hh
}

// ParseStruct parse struct into header
func ParseStruct(h Header, v interface{}) Header {
	data, err := json.Marshal(v)
	if err != nil {
		return h
	}

	err = json.Unmarshal(data, &h)
	return h
}

// HeaderFromStruct init header from struct
func HeaderFromStruct(v interface{}) Header {

	var header Header
	header = ParseStruct(header, v)
	return header
}

type ReservedHeader map[string]string
