package req

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestToJSON(t *testing.T) {
	type Result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	r1 := Result{
		Code: 1,
		Msg:  "ok",
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(&r1)
		w.Write(data)
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	r, err := Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	var r2 Result
	err = r.ToJSON(&r2)
	if err != nil {
		t.Fatal(err)
	}
	if r1 != r2 {
		t.Errorf("json response body = %+v; want = %+v", r2, r1)
	}
}

func TestToXML(t *testing.T) {
	type Result struct {
		XMLName xml.Name
		Code    int    `xml:"code"`
		Msg     string `xml:"msg"`
	}
	r1 := Result{
		XMLName: xml.Name{Local: "result"},
		Code:    1,
		Msg:     "ok",
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		data, _ := xml.Marshal(&r1)
		w.Write(data)
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	r, err := Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	var r2 Result
	err = r.ToXML(&r2)
	if err != nil {
		t.Fatal(err)
	}
	if r1 != r2 {
		t.Errorf("xml response body = %+v; want = %+v", r2, r1)
	}
}

func TestFormat(t *testing.T) {
	SetFlags(LstdFlags | Lcost)
	reqHeader := "Request-Header"
	respHeader := "Response-Header"
	reqBody := "request body"
	respBody1 := "response body 1"
	respBody2 := "response body 2"
	respBody := fmt.Sprintf("%s\n%s", respBody1, respBody2)
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(respHeader, "req")
		w.Write([]byte(respBody))
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))

	// %v
	r, err := Post(ts.URL, reqBody, Header{reqHeader: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	str := fmt.Sprintf("%v", r)
	for _, keyword := range []string{ts.URL, reqBody, respBody} {
		if !strings.Contains(str, keyword) {
			t.Errorf("format %%v output lack of part, want: %s", keyword)
		}
	}

	// %-v
	str = fmt.Sprintf("%-v", r)
	for _, keyword := range []string{ts.URL, respBody1 + " " + respBody2} {
		if !strings.Contains(str, keyword) {
			t.Errorf("format %%-v output lack of part, want: %s", keyword)
		}
	}

	// %+v
	str = fmt.Sprintf("%+v", r)
	for _, keyword := range []string{reqBody, respBody, reqHeader, respHeader} {
		if !strings.Contains(str, keyword) {
			t.Errorf("format %%+v output lack of part, want: %s", keyword)
		}
	}
}

func TestBytesAndString(t *testing.T) {
	respBody := "response body"
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(respBody))
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	r, err := Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(r.Bytes()) != respBody {
		t.Errorf("response body = %s; want = %s", r.Bytes(), respBody)
	}
	if r.String() != respBody {
		t.Errorf("response body = %s; want = %s", r.String(), respBody)
	}
}
