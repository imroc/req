package req

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Body struct {
	io.ReadCloser
	resp *http.Response
}

func (body Body) MustSave(dst io.Writer) {
	err := body.Save(dst)
	if err != nil {
		panic(err)
	}
}

func (body Body) Save(dst io.Writer) error {
	if dst == nil {
		return nil // TODO: return error
	}
	_, err := io.Copy(dst, body.ReadCloser)
	body.Close()
	return err
}

func (body Body) MustSaveFile(filename string) {
	err := body.SaveFile(filename)
	if err != nil {
		panic(err)
	}
}

func (body Body) SaveFile(filename string) error {
	if filename == "" {
		return nil // TODO: return error
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, body.ReadCloser)
	body.Close()
	return err
}

func (body Body) MustUnmarshalJson(v interface{}) {
	err := body.UnmarshalJson(v)
	if err != nil {
		panic(err)
	}
}

func (body Body) UnmarshalJson(v interface{}) error {
	b, err := body.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func (body Body) MustUnmarshalXml(v interface{}) {
	err := body.UnmarshalXml(v)
	if err != nil {
		panic(err)
	}
}

func (body Body) UnmarshalXml(v interface{}) error {
	b, err := body.Bytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(b, v)
}
func (body Body) MustUnmarshal(v interface{}) {
	err := body.Unmarshal(v)
	if err != nil {
		panic(err)
	}
}

func (body Body) Unmarshal(v interface{}) error {
	contentType := body.resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "json") {
		return body.UnmarshalJson(v)
	} else if strings.Contains(contentType, "xml") {
		return body.UnmarshalXml(v)
	}
	return body.UnmarshalJson(v)
}

func (body Body) MustString() string {
	b, err := body.Bytes()
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (body Body) String() (string, error) {
	b, err := body.Bytes()
	return string(b), err
}

func (body Body) Bytes() ([]byte, error) {
	defer body.Close()
	return ioutil.ReadAll(body.ReadCloser)
}

func (body Body) MustBytes() []byte {
	b, err := body.Bytes()
	if err != nil {
		panic(err)
	}
	return b
}
