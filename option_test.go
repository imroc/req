package req

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBodyJSON(t *testing.T) {
	type content struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	c := content{
		Code: 1,
		Msg:  "ok",
	}
	checkData := func(data []byte) {
		var cc content
		err := json.Unmarshal(data, &cc)
		if err != nil {
			t.Fatalf("error unmarshal request body in BodyJSON, json.Unmarshal: %v", err)
		}
		if cc != c {
			t.Errorf("request body = %+v; want = %+v", cc, c)
		}
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("error read request body in BodyJSON, ReadAll: %v", err)
		}
		checkData(data)
	})
	ts := httptest.NewServer(handler)

	resp, err := Post(ts.URL, BodyJSON(&c))
	if err != nil {
		t.Fatalf("error initiate request in BodyJSON, Post: %v", err)
	}
	checkData(resp.reqBody)

	SetJSONEscapeHTML(false)
	SetJSONIndent("", "\t")
	resp, err = Put(ts.URL, BodyJSON(&c))
	if err != nil {
		t.Fatalf("error initiate request in BodyJSON, Put: %v", err)
	}
	checkData(resp.reqBody)
}

func TestBodyXML(t *testing.T) {
	type content struct {
		Code int    `xml:"code"`
		Msg  string `xml:"msg"`
	}
	c := content{
		Code: 1,
		Msg:  "ok",
	}
	checkData := func(data []byte) {
		var cc content
		err := xml.Unmarshal(data, &cc)
		if err != nil {
			t.Fatalf("error unmarshal request body in BodyXML, xml.Unmarshal: %v", err)
		}
		if cc != c {
			t.Errorf("request body = %+v; want = %+v", cc, c)
		}
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("error read request body in BodyXML, ReadAll: %v", err)
		}
		checkData(data)
	})
	ts := httptest.NewServer(handler)

	resp, err := Post(ts.URL, BodyXML(&c))
	if err != nil {
		t.Fatalf("error initiate request in BodyXML, Post: %v", err)
	}
	checkData(resp.reqBody)

	SetXMLIndent("", "    ")
	resp, err = Put(ts.URL, BodyXML(&c))
	if err != nil {
		t.Fatalf("error initiate request in BodyXML, Put: %v", err)
	}
	checkData(resp.reqBody)
}
