package req

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"
)

func init() {
	SetLogger(nil) // disable log
}

func assertRequestNotNil(t *testing.T, rs ...*Request) {
	for _, r := range rs {
		assertNotNil(t, r)
	}
}

func TestGlobalWrapperForRequestSettings(t *testing.T) {
	assertRequestNotNil(t,
		SetFiles(map[string]string{"test": "test"}),
		SetFile("test", "test"),
		SetFileReader("test", "test.txt", bytes.NewBufferString("test")),
		SetFileBytes("test", "test.txt", []byte("test")),
		SetFileUpload(FileUpload{}),
		SetError(&ErrorMessage{}),
		SetResult(&UserInfo{}),
		SetOutput(nil),
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
		SetFormDataFromValues(nil),
		SetContentType(jsonContentType),
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
	)
}
