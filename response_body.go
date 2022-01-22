package req

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func (r *Response) MustSave(dst io.Writer) {
	err := r.Save(dst)
	if err != nil {
		panic(err)
	}
}

func (r *Response) Save(dst io.Writer) error {
	if dst == nil {
		return nil // TODO: return error
	}
	_, err := io.Copy(dst, r.Body)
	r.Body.Close()
	return err
}

func (r *Response) MustSaveFile(filename string) {
	err := r.SaveFile(filename)
	if err != nil {
		panic(err)
	}
}

func (r *Response) SaveFile(filename string) error {
	if filename == "" {
		return nil // TODO: return error
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, r.Body)
	r.Body.Close()
	return err
}

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
	return ioutil.ReadAll(r.Body)
}

func (r *Response) MustBytes() []byte {
	b, err := r.Bytes()
	if err != nil {
		panic(err)
	}
	return b
}
