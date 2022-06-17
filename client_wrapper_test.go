package req

import (
	"crypto/tls"
	"github.com/imroc/req/v3/internal/tests"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGlobalWrapper(t *testing.T) {
	EnableInsecureSkipVerify()
	testGlobalWrapperSendMethods(t)
	testGlobalWrapperMustSendMethods(t)
	DisableInsecureSkipVerify()

	u, _ := url.Parse("http://dummy.proxy.local")
	proxy := http.ProxyURL(u)
	form := make(url.Values)
	form.Add("test", "test")

	tests.AssertAllNotNil(t,
		SetCommonError(nil),
		SetCookieJar(nil),
		SetDialTLS(nil),
		SetDial(nil),
		SetTLSHandshakeTimeout(time.Second),
		EnableAllowGetMethodPayload(),
		DisableAllowGetMethodPayload(),
		SetJsonMarshal(nil),
		SetJsonUnmarshal(nil),
		SetXmlMarshal(nil),
		SetXmlUnmarshal(nil),
		EnableTraceAll(),
		DisableTraceAll(),
		OnAfterResponse(func(client *Client, response *Response) error {
			return nil
		}),
		OnBeforeRequest(func(client *Client, request *Request) error {
			return nil
		}),
		SetProxyURL("http://dummy.proxy.local"),
		SetProxyURL("bad url"),
		SetProxy(proxy),
		SetCommonContentType(jsonContentType),
		SetCommonHeader("my-header", "my-value"),
		SetCommonHeaders(map[string]string{
			"header1": "value1",
			"header2": "value2",
		}),
		SetCommonBasicAuth("imroc", "123456"),
		SetCommonBearerAuthToken("123456"),
		SetUserAgent("test"),
		SetTimeout(1*time.Second),
		SetLogger(createDefaultLogger()),
		SetScheme("https"),
		EnableDebugLog(),
		DisableDebugLog(),
		SetCommonCookies(&http.Cookie{Name: "test", Value: "test"}),
		SetCommonQueryString("test1=test1"),
		SetCommonPathParams(map[string]string{"test1": "test1"}),
		SetCommonPathParam("test2", "test2"),
		AddCommonQueryParam("test1", "test11"),
		SetCommonQueryParam("test1", "test111"),
		SetCommonQueryParams(map[string]string{"test1": "test1"}),
		EnableInsecureSkipVerify(),
		DisableInsecureSkipVerify(),
		DisableCompression(),
		EnableCompression(),
		DisableKeepAlives(),
		EnableKeepAlives(),
		SetRootCertsFromFile(tests.GetTestFilePath("sample-root.pem")),
		SetRootCertFromString(string(getTestFileContent(t, "sample-root.pem"))),
		SetCerts(tls.Certificate{}, tls.Certificate{}),
		SetCertFromFile(
			tests.GetTestFilePath("sample-client.pem"),
			tests.GetTestFilePath("sample-client-key.pem"),
		),
		SetOutputDirectory(testDataPath),
		SetBaseURL("http://dummy-req.local/test"),
		SetCommonFormDataFromValues(form),
		SetCommonFormData(map[string]string{"test2": "test2"}),
		DisableAutoReadResponse(),
		EnableAutoReadResponse(),
		EnableDumpAll(),
		EnableDumpAllAsync(),
		EnableDumpAllWithoutBody(),
		EnableDumpAllWithoutResponse(),
		EnableDumpAllWithoutRequest(),
		EnableDumpAllWithoutHeader(),
		SetLogger(nil),
		EnableDumpAllToFile(filepath.Join(testDataPath, "path-not-exists", "dump.out")),
		EnableDumpAllToFile(tests.GetTestFilePath("tmpdump.out")),
		SetCommonDumpOptions(&DumpOptions{
			RequestHeader: true,
		}),
		DisableDumpAll(),
		SetRedirectPolicy(NoRedirectPolicy()),
		EnableForceHTTP1(),
		EnableForceHTTP2(),
		DisableForceHttpVersion(),
		SetAutoDecodeContentType("json"),
		SetAutoDecodeContentTypeFunc(func(contentType string) bool { return true }),
		SetAutoDecodeAllContentType(),
		DisableAutoDecode(),
		EnableAutoDecode(),
		AddCommonRetryCondition(func(resp *Response, err error) bool { return true }),
		SetCommonRetryCondition(func(resp *Response, err error) bool { return true }),
		AddCommonRetryHook(func(resp *Response, err error) {}),
		SetCommonRetryHook(func(resp *Response, err error) {}),
		SetCommonRetryCount(2),
		SetCommonRetryInterval(func(resp *Response, attempt int) time.Duration {
			return 1 * time.Second
		}),
		SetCommonRetryBackoffInterval(1*time.Millisecond, 2*time.Second),
		SetCommonRetryFixedInterval(1*time.Second),
		SetUnixSocket("/var/run/custom.sock"),
	)
	os.Remove(tests.GetTestFilePath("tmpdump.out"))

	config := GetTLSClientConfig()
	tests.AssertEqual(t, config, DefaultClient().t.TLSClientConfig)

	r := R()
	tests.AssertEqual(t, true, r != nil)
	c := C()

	c.SetTimeout(10 * time.Second)
	SetDefaultClient(c)
	tests.AssertEqual(t, true, DefaultClient().httpClient.Timeout == 10*time.Second)
	tests.AssertEqual(t, GetClient(), DefaultClient().httpClient)

	r = NewRequest()
	tests.AssertEqual(t, true, r != nil)
	c = NewClient()
	tests.AssertEqual(t, true, c != nil)
}
