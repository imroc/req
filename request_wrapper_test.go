package req

import (
	"bytes"
	"context"
	"github.com/imroc/req/v3/internal/header"
	"github.com/imroc/req/v3/internal/tests"
	"net/http"
	"testing"
	"time"
)

func init() {
	SetLogger(nil) // disable log
}

func TestGlobalWrapperForRequestSettings(t *testing.T) {
	tests.AssertAllNotNil(t,
		SetFiles(map[string]string{"test": "req.go"}),
		SetFile("test", "req.go"),
		SetFileReader("test", "test.txt", bytes.NewBufferString("test")),
		SetFileBytes("test", "test.txt", []byte("test")),
		SetFileUpload(FileUpload{}),
		SetError(&ErrorMessage{}),
		SetResult(&UserInfo{}),
		SetOutput(new(bytes.Buffer)),
		SetHeader("test", "test"),
		SetHeaders(map[string]string{"test": "test"}),
		SetCookies(&http.Cookie{
			Name:  "test",
			Value: "test",
		}),
		SetBasicAuth("imroc", "123456"),
		SetBearerAuthToken("123456"),
		SetQueryString("test=test"),
		SetQueryString("ksjlfjk?"),
		SetQueryParam("test", "test"),
		AddQueryParam("test", "test"),
		SetQueryParams(map[string]string{"test": "test"}),
		SetPathParam("test", "test"),
		SetPathParams(map[string]string{"test": "test"}),
		SetFormData(map[string]string{"test": "test"}),
		SetURL(""),
		SetFormDataFromValues(nil),
		SetContentType(header.JsonContentType),
		AddRetryCondition(func(rep *Response, err error) bool {
			return err != nil
		}),
		SetRetryCondition(func(rep *Response, err error) bool {
			return err != nil
		}),
		AddRetryHook(func(resp *Response, err error) {}),
		SetRetryHook(func(resp *Response, err error) {}),
		SetRetryBackoffInterval(0, 0),
		SetRetryFixedInterval(0),
		SetRetryInterval(func(resp *Response, attempt int) time.Duration {
			return 1 * time.Millisecond
		}),
		SetRetryCount(3),
		SetBodyXmlMarshal(0),
		SetBodyString("test"),
		SetBodyBytes([]byte("test")),
		SetBodyJsonBytes([]byte(`{"user":"roc"}`)),
		SetBodyJsonString(`{"user":"roc"}`),
		SetBodyXmlBytes([]byte("test")),
		SetBodyXmlString("test"),
		SetBody("test"),
		SetBodyJsonMarshal(User{
			Name: "roc",
		}),
		EnableTrace(),
		DisableTrace(),
		SetContext(context.Background()),
		SetUploadCallback(nil),
		SetUploadCallbackWithInterval(nil, 0),
		SetDownloadCallback(nil),
		SetDownloadCallbackWithInterval(nil, 0),
	)
}

func testGlobalWrapperMustSendMethods(t *testing.T) {
	testCases := []struct {
		SendReq      func(string) *Response
		ExpectMethod string
	}{
		{
			SendReq:      MustGet,
			ExpectMethod: "GET",
		},
		{
			SendReq:      MustPost,
			ExpectMethod: "POST",
		},
		{
			SendReq:      MustPatch,
			ExpectMethod: "PATCH",
		},
		{
			SendReq:      MustPut,
			ExpectMethod: "PUT",
		},
		{
			SendReq:      MustDelete,
			ExpectMethod: "DELETE",
		},
		{
			SendReq:      MustOptions,
			ExpectMethod: "OPTIONS",
		},
		{
			SendReq:      MustHead,
			ExpectMethod: "HEAD",
		},
	}
	url := getTestServerURL() + "/"
	for _, tc := range testCases {
		resp := tc.SendReq(url)
		tests.AssertNotNil(t, resp.Response)
		tests.AssertEqual(t, tc.ExpectMethod, resp.Header.Get("Method"))
	}
}

func testGlobalWrapperSendMethods(t *testing.T) {
	testCases := []struct {
		SendReq      func(string) (*Response, error)
		ExpectMethod string
	}{
		{
			SendReq:      Get,
			ExpectMethod: "GET",
		},
		{
			SendReq:      Post,
			ExpectMethod: "POST",
		},
		{
			SendReq:      Patch,
			ExpectMethod: "PATCH",
		},
		{
			SendReq:      Put,
			ExpectMethod: "PUT",
		},
		{
			SendReq:      Delete,
			ExpectMethod: "DELETE",
		},
		{
			SendReq:      Options,
			ExpectMethod: "OPTIONS",
		},
		{
			SendReq:      Head,
			ExpectMethod: "HEAD",
		},
	}
	url := getTestServerURL() + "/"
	for _, tc := range testCases {
		resp, err := tc.SendReq(url)
		assertSuccess(t, resp, err)
		tests.AssertEqual(t, tc.ExpectMethod, resp.Header.Get("Method"))
	}
}
