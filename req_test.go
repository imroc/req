package req

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUrlParam(t *testing.T) {
	m := map[string]interface{}{
		"access_token": "123abc",
		"name":         "roc",
		"enc":          "中文",
	}
	queryHandler := func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		for key, value := range m {
			if v := query.Get(key); value != v {
				t.Errorf("query param %s = %s; want = %s", key, v, value)
			}
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(queryHandler))
	_, err := Get(ts.URL, QueryParam(m))
	if err != nil {
		t.Fatal(err)
	}
	_, err = Head(ts.URL, Param(m))
	if err != nil {
		t.Fatal(err)
	}
	_, err = Put(ts.URL, QueryParam(m))
	if err != nil {
		t.Fatal(err)
	}
}

func TestFormParam(t *testing.T) {
	formParam := Param{
		"access_token": "123abc",
		"name":         "roc",
		"enc":          "中文",
	}
	formHandler := func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		for key, value := range formParam {
			if v := r.FormValue(key); value != v {
				t.Errorf("form param %s = %s; want = %s", key, v, value)
			}
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(formHandler))
	url := ts.URL
	_, err := Post(url, formParam)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParamWithBody(t *testing.T) {
	reqBody := "request body"
	p := Param{
		"name": "roc",
		"job":  "programmer",
	}
	buf := bytes.NewBufferString(reqBody)
	ts := newDefaultTestServer()
	r, err := Post(ts.URL, p, buf)
	if err != nil {
		t.Fatal(err)
	}
	if r.Request().URL.Query().Get("name") != "roc" {
		t.Error("param should in the url when set body manually")
	}
	if string(r.reqBody) != reqBody {
		t.Error("request body not equal")
	}
}

func TestParamBoth(t *testing.T) {
	urlParam := QueryParam{
		"access_token": "123abc",
		"enc":          "中文",
	}
	formParam := Param{
		"name": "roc",
		"job":  "软件工程师",
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		for key, value := range urlParam {
			if v := query.Get(key); value != v {
				t.Errorf("query param %s = %s; want = %s", key, v, value)
			}
		}
		r.ParseForm()
		for key, value := range formParam {
			if v := r.FormValue(key); value != v {
				t.Errorf("form param %s = %s; want = %s", key, v, value)
			}
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	url := ts.URL
	_, err := Patch(url, urlParam, formParam)
	if err != nil {
		t.Fatal(err)
	}

}

func TestBody(t *testing.T) {
	body := "request body"
	handler := func(w http.ResponseWriter, r *http.Request) {
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(bs) != body {
			t.Errorf("body = %s; want = %s", bs, body)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))

	// string
	_, err := Post(ts.URL, body)
	if err != nil {
		t.Fatal(err)
	}

	// []byte
	_, err = Post(ts.URL, []byte(body))
	if err != nil {
		t.Fatal(err)
	}

	// *bytes.Buffer
	var buf bytes.Buffer
	buf.WriteString(body)
	_, err = Post(ts.URL, &buf)
	if err != nil {
		t.Fatal(err)
	}

	// io.Reader
	_, err = Post(ts.URL, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
}

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
			t.Fatal(err)
		}
		if cc != c {
			t.Errorf("request body = %+v; want = %+v", cc, c)
		}
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		checkData(data)
	})

	ts := httptest.NewServer(handler)
	resp, err := Post(ts.URL, BodyJSON(&c))
	if err != nil {
		t.Fatal(err)
	}
	checkData(resp.reqBody)

	SetJSONEscapeHTML(false)
	SetJSONIndent("", "\t")
	resp, err = Put(ts.URL, BodyJSON(&c))
	if err != nil {
		t.Fatal(err)
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
			t.Fatal(err)
		}
		if cc != c {
			t.Errorf("request body = %+v; want = %+v", cc, c)
		}
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		checkData(data)
	})

	ts := httptest.NewServer(handler)
	resp, err := Post(ts.URL, BodyXML(&c))
	if err != nil {
		t.Fatal(err)
	}
	checkData(resp.reqBody)

	SetXMLIndent("", "    ")
	resp, err = Put(ts.URL, BodyXML(&c))
	if err != nil {
		t.Fatal(err)
	}
	checkData(resp.reqBody)
}

func TestHeader(t *testing.T) {
	header := Header{
		"User-Agent":    "V1.0.0",
		"Authorization": "roc",
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		for key, value := range header {
			if v := r.Header.Get(key); value != v {
				t.Errorf("header %q = %s; want = %s", key, v, value)
			}
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	_, err := Head(ts.URL, header)
	if err != nil {
		t.Fatal(err)
	}

	httpHeader := make(http.Header)
	for key, value := range header {
		httpHeader.Add(key, value)
	}
	_, err = Head(ts.URL, httpHeader)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpload(t *testing.T) {
	str := "hello req"
	file := ioutil.NopCloser(strings.NewReader(str))
	upload := FileUpload{
		File:      file,
		FieldName: "media",
		FileName:  "hello.txt",
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		mr, err := r.MultipartReader()
		if err != nil {
			t.Fatal(err)
		}
		for {
			p, err := mr.NextPart()
			if err != nil {
				break
			}
			if p.FileName() != upload.FileName {
				t.Errorf("filename = %s; want = %s", p.FileName(), upload.FileName)
			}
			if p.FormName() != upload.FieldName {
				t.Errorf("formname = %s; want = %s", p.FileName(), upload.FileName)
			}
			data, err := ioutil.ReadAll(p)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != str {
				t.Errorf("file content = %s; want = %s", data, str)
			}
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	_, err := Post(ts.URL, upload)
	if err != nil {
		t.Fatal(err)
	}
	ts = newDefaultTestServer()
	_, err = Post(ts.URL, File("*.go"))
	if err != nil {
		t.Fatal(err)
	}
}
