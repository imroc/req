package req

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"strings"
)

func (r *Response) MustUnmarshalJson(v interface{}) {
	err := r.UnmarshalJson(v)
	if err != nil {
		panic(err)
	}
}

func (r *Response) UnmarshalJson(v interface{}) error {
	b, err := r.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func (r *Response) MustUnmarshalXml(v interface{}) {
	err := r.UnmarshalXml(v)
	if err != nil {
		panic(err)
	}
}

func (r *Response) UnmarshalXml(v interface{}) error {
	b, err := r.Bytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(b, v)
}
func (r *Response) MustUnmarshal(v interface{}) {
	err := r.Unmarshal(v)
	if err != nil {
		panic(err)
	}
}

func (r *Response) Unmarshal(v interface{}) error {
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "json") {
		return r.UnmarshalJson(v)
	} else if strings.Contains(contentType, "xml") {
		return r.UnmarshalXml(v)
	}
	return r.UnmarshalJson(v)
}

func (r *Response) MustString() string {
	b, err := r.Bytes()
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (r *Response) String() (string, error) {
	b, err := r.Bytes()
	return string(b), err
}

func (r *Response) Bytes() ([]byte, error) {
	if r.body != nil {
		return r.body, nil
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.body = body
	return body, nil
}

func (r *Response) MustBytes() []byte {
	b, err := r.Bytes()
	if err != nil {
		panic(err)
	}
	return b
}
