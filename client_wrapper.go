package req

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// WrapRoundTrip is a global wrapper methods which delegated
// to the default client's WrapRoundTrip.
func WrapRoundTrip(wrappers ...RoundTripWrapper) *Client {
	return defaultClient.WrapRoundTrip(wrappers...)
}

// WrapRoundTripFunc is a global wrapper methods which delegated
// to the default client's WrapRoundTripFunc.
func WrapRoundTripFunc(funcs ...RoundTripWrapperFunc) *Client {
	return defaultClient.WrapRoundTripFunc(funcs...)
}

// SetCommonError is a global wrapper methods which delegated
// to the default client's SetCommonErrorResult.
//
// Deprecated: Use SetCommonErrorResult instead.
func SetCommonError(err interface{}) *Client {
	return defaultClient.SetCommonErrorResult(err)
}

// SetCommonErrorResult is a global wrapper methods which delegated
// to the default client's SetCommonError.
func SetCommonErrorResult(err interface{}) *Client {
	return defaultClient.SetCommonErrorResult(err)
}

// SetCommonUnknownResultHandlerFunc is a global wrapper methods which delegated
// to the default client's SetCommonUnknownResultHandlerFunc.
func SetCommonUnknownResultHandlerFunc(fn func(resp *Response) error) *Client {
	return defaultClient.SetCommonUnknownResultHandlerFunc(fn)
}

// SetCommonResultStateCheckFunc is a global wrapper methods which delegated
// to the default client's SetCommonResultStateCheckFunc.
func SetCommonResultStateCheckFunc(fn func(resp *Response) ResultState) *Client {
	return defaultClient.SetCommonResultStateCheckFunc(fn)
}

// SetCommonFormDataFromValues is a global wrapper methods which delegated
// to the default client's SetCommonFormDataFromValues.
func SetCommonFormDataFromValues(data url.Values) *Client {
	return defaultClient.SetCommonFormDataFromValues(data)
}

// SetCommonFormData is a global wrapper methods which delegated
// to the default client's SetCommonFormData.
func SetCommonFormData(data map[string]string) *Client {
	return defaultClient.SetCommonFormData(data)
}

// SetBaseURL is a global wrapper methods which delegated
// to the default client's SetBaseURL.
func SetBaseURL(u string) *Client {
	return defaultClient.SetBaseURL(u)
}

// SetOutputDirectory is a global wrapper methods which delegated
// to the default client's SetOutputDirectory.
func SetOutputDirectory(dir string) *Client {
	return defaultClient.SetOutputDirectory(dir)
}

// SetCertFromFile is a global wrapper methods which delegated
// to the default client's SetCertFromFile.
func SetCertFromFile(certFile, keyFile string) *Client {
	return defaultClient.SetCertFromFile(certFile, keyFile)
}

// SetCerts is a global wrapper methods which delegated
// to the default client's SetCerts.
func SetCerts(certs ...tls.Certificate) *Client {
	return defaultClient.SetCerts(certs...)
}

// SetRootCertFromString is a global wrapper methods which delegated
// to the default client's SetRootCertFromString.
func SetRootCertFromString(pemContent string) *Client {
	return defaultClient.SetRootCertFromString(pemContent)
}

// SetRootCertsFromFile is a global wrapper methods which delegated
// to the default client's SetRootCertsFromFile.
func SetRootCertsFromFile(pemFiles ...string) *Client {
	return defaultClient.SetRootCertsFromFile(pemFiles...)
}

// GetTLSClientConfig is a global wrapper methods which delegated
// to the default client's GetTLSClientConfig.
func GetTLSClientConfig() *tls.Config {
	return defaultClient.GetTLSClientConfig()
}

// SetRedirectPolicy is a global wrapper methods which delegated
// to the default client's SetRedirectPolicy.
func SetRedirectPolicy(policies ...RedirectPolicy) *Client {
	return defaultClient.SetRedirectPolicy(policies...)
}

// DisableKeepAlives is a global wrapper methods which delegated
// to the default client's DisableKeepAlives.
func DisableKeepAlives() *Client {
	return defaultClient.DisableKeepAlives()
}

// EnableKeepAlives is a global wrapper methods which delegated
// to the default client's EnableKeepAlives.
func EnableKeepAlives() *Client {
	return defaultClient.EnableKeepAlives()
}

// DisableCompression is a global wrapper methods which delegated
// to the default client's DisableCompression.
func DisableCompression() *Client {
	return defaultClient.DisableCompression()
}

// EnableCompression is a global wrapper methods which delegated
// to the default client's EnableCompression.
func EnableCompression() *Client {
	return defaultClient.EnableCompression()
}

// SetTLSClientConfig is a global wrapper methods which delegated
// to the default client's SetTLSClientConfig.
func SetTLSClientConfig(conf *tls.Config) *Client {
	return defaultClient.SetTLSClientConfig(conf)
}

// EnableInsecureSkipVerify is a global wrapper methods which delegated
// to the default client's EnableInsecureSkipVerify.
func EnableInsecureSkipVerify() *Client {
	return defaultClient.EnableInsecureSkipVerify()
}

// DisableInsecureSkipVerify is a global wrapper methods which delegated
// to the default client's DisableInsecureSkipVerify.
func DisableInsecureSkipVerify() *Client {
	return defaultClient.DisableInsecureSkipVerify()
}

// SetCommonQueryParams is a global wrapper methods which delegated
// to the default client's SetCommonQueryParams.
func SetCommonQueryParams(params map[string]string) *Client {
	return defaultClient.SetCommonQueryParams(params)
}

// AddCommonQueryParam is a global wrapper methods which delegated
// to the default client's AddCommonQueryParam.
func AddCommonQueryParam(key, value string) *Client {
	return defaultClient.AddCommonQueryParam(key, value)
}

// AddCommonQueryParams is a global wrapper methods which delegated
// to the default client's AddCommonQueryParams.
func AddCommonQueryParams(key string, values ...string) *Client {
	return defaultClient.AddCommonQueryParams(key, values...)
}

// SetCommonPathParam is a global wrapper methods which delegated
// to the default client's SetCommonPathParam.
func SetCommonPathParam(key, value string) *Client {
	return defaultClient.SetCommonPathParam(key, value)
}

// SetCommonPathParams is a global wrapper methods which delegated
// to the default client's SetCommonPathParams.
func SetCommonPathParams(pathParams map[string]string) *Client {
	return defaultClient.SetCommonPathParams(pathParams)
}

// SetCommonQueryParam is a global wrapper methods which delegated
// to the default client's SetCommonQueryParam.
func SetCommonQueryParam(key, value string) *Client {
	return defaultClient.SetCommonQueryParam(key, value)
}

// SetCommonQueryString is a global wrapper methods which delegated
// to the default client's SetCommonQueryString.
func SetCommonQueryString(query string) *Client {
	return defaultClient.SetCommonQueryString(query)
}

// SetCommonCookies is a global wrapper methods which delegated
// to the default client's SetCommonCookies.
func SetCommonCookies(cookies ...*http.Cookie) *Client {
	return defaultClient.SetCommonCookies(cookies...)
}

// DisableDebugLog is a global wrapper methods which delegated
// to the default client's DisableDebugLog.
func DisableDebugLog() *Client {
	return defaultClient.DisableDebugLog()
}

// EnableDebugLog is a global wrapper methods which delegated
// to the default client's EnableDebugLog.
func EnableDebugLog() *Client {
	return defaultClient.EnableDebugLog()
}

// DevMode is a global wrapper methods which delegated
// to the default client's DevMode.
func DevMode() *Client {
	return defaultClient.DevMode()
}

// SetScheme is a global wrapper methods which delegated
// to the default client's SetScheme.
func SetScheme(scheme string) *Client {
	return defaultClient.SetScheme(scheme)
}

// SetLogger is a global wrapper methods which delegated
// to the default client's SetLogger.
func SetLogger(log Logger) *Client {
	return defaultClient.SetLogger(log)
}

// SetTimeout is a global wrapper methods which delegated
// to the default client's SetTimeout.
func SetTimeout(d time.Duration) *Client {
	return defaultClient.SetTimeout(d)
}

// EnableDumpAll is a global wrapper methods which delegated
// to the default client's EnableDumpAll.
func EnableDumpAll() *Client {
	return defaultClient.EnableDumpAll()
}

// EnableDumpAllToFile is a global wrapper methods which delegated
// to the default client's EnableDumpAllToFile.
func EnableDumpAllToFile(filename string) *Client {
	return defaultClient.EnableDumpAllToFile(filename)
}

// EnableDumpAllTo is a global wrapper methods which delegated
// to the default client's EnableDumpAllTo.
func EnableDumpAllTo(output io.Writer) *Client {
	return defaultClient.EnableDumpAllTo(output)
}

// EnableDumpAllAsync is a global wrapper methods which delegated
// to the default client's EnableDumpAllAsync.
func EnableDumpAllAsync() *Client {
	return defaultClient.EnableDumpAllAsync()
}

// EnableDumpAllWithoutRequestBody is a global wrapper methods which delegated
// to the default client's EnableDumpAllWithoutRequestBody.
func EnableDumpAllWithoutRequestBody() *Client {
	return defaultClient.EnableDumpAllWithoutRequestBody()
}

// EnableDumpAllWithoutResponseBody is a global wrapper methods which delegated
// to the default client's EnableDumpAllWithoutResponseBody.
func EnableDumpAllWithoutResponseBody() *Client {
	return defaultClient.EnableDumpAllWithoutResponseBody()
}

// EnableDumpAllWithoutResponse is a global wrapper methods which delegated
// to the default client's EnableDumpAllWithoutResponse.
func EnableDumpAllWithoutResponse() *Client {
	return defaultClient.EnableDumpAllWithoutResponse()
}

// EnableDumpAllWithoutRequest is a global wrapper methods which delegated
// to the default client's EnableDumpAllWithoutRequest.
func EnableDumpAllWithoutRequest() *Client {
	return defaultClient.EnableDumpAllWithoutRequest()
}

// EnableDumpAllWithoutHeader is a global wrapper methods which delegated
// to the default client's EnableDumpAllWithoutHeader.
func EnableDumpAllWithoutHeader() *Client {
	return defaultClient.EnableDumpAllWithoutHeader()
}

// EnableDumpAllWithoutBody is a global wrapper methods which delegated
// to the default client's EnableDumpAllWithoutBody.
func EnableDumpAllWithoutBody() *Client {
	return defaultClient.EnableDumpAllWithoutBody()
}

// EnableDumpEachRequest is a global wrapper methods which delegated
// to the default client's EnableDumpEachRequest.
func EnableDumpEachRequest() *Client {
	return defaultClient.EnableDumpEachRequest()
}

// EnableDumpEachRequestWithoutBody is a global wrapper methods which delegated
// to the default client's EnableDumpEachRequestWithoutBody.
func EnableDumpEachRequestWithoutBody() *Client {
	return defaultClient.EnableDumpEachRequestWithoutBody()
}

// EnableDumpEachRequestWithoutHeader is a global wrapper methods which delegated
// to the default client's EnableDumpEachRequestWithoutHeader.
func EnableDumpEachRequestWithoutHeader() *Client {
	return defaultClient.EnableDumpEachRequestWithoutHeader()
}

// EnableDumpEachRequestWithoutResponse is a global wrapper methods which delegated
// to the default client's EnableDumpEachRequestWithoutResponse.
func EnableDumpEachRequestWithoutResponse() *Client {
	return defaultClient.EnableDumpEachRequestWithoutResponse()
}

// EnableDumpEachRequestWithoutRequest is a global wrapper methods which delegated
// to the default client's EnableDumpEachRequestWithoutRequest.
func EnableDumpEachRequestWithoutRequest() *Client {
	return defaultClient.EnableDumpEachRequestWithoutRequest()
}

// EnableDumpEachRequestWithoutResponseBody is a global wrapper methods which delegated
// to the default client's EnableDumpEachRequestWithoutResponseBody.
func EnableDumpEachRequestWithoutResponseBody() *Client {
	return defaultClient.EnableDumpEachRequestWithoutResponseBody()
}

// EnableDumpEachRequestWithoutRequestBody is a global wrapper methods which delegated
// to the default client's EnableDumpEachRequestWithoutRequestBody.
func EnableDumpEachRequestWithoutRequestBody() *Client {
	return defaultClient.EnableDumpEachRequestWithoutRequestBody()
}

// DisableAutoReadResponse is a global wrapper methods which delegated
// to the default client's DisableAutoReadResponse.
func DisableAutoReadResponse() *Client {
	return defaultClient.DisableAutoReadResponse()
}

// EnableAutoReadResponse is a global wrapper methods which delegated
// to the default client's EnableAutoReadResponse.
func EnableAutoReadResponse() *Client {
	return defaultClient.EnableAutoReadResponse()
}

// SetAutoDecodeContentType is a global wrapper methods which delegated
// to the default client's SetAutoDecodeContentType.
func SetAutoDecodeContentType(contentTypes ...string) *Client {
	return defaultClient.SetAutoDecodeContentType(contentTypes...)
}

// SetAutoDecodeContentTypeFunc is a global wrapper methods which delegated
// to the default client's SetAutoDecodeAllTypeFunc.
func SetAutoDecodeContentTypeFunc(fn func(contentType string) bool) *Client {
	return defaultClient.SetAutoDecodeContentTypeFunc(fn)
}

// SetAutoDecodeAllContentType is a global wrapper methods which delegated
// to the default client's SetAutoDecodeAllContentType.
func SetAutoDecodeAllContentType() *Client {
	return defaultClient.SetAutoDecodeAllContentType()
}

// DisableAutoDecode is a global wrapper methods which delegated
// to the default client's DisableAutoDecode.
func DisableAutoDecode() *Client {
	return defaultClient.DisableAutoDecode()
}

// EnableAutoDecode is a global wrapper methods which delegated
// to the default client's EnableAutoDecode.
func EnableAutoDecode() *Client {
	return defaultClient.EnableAutoDecode()
}

// SetUserAgent is a global wrapper methods which delegated
// to the default client's SetUserAgent.
func SetUserAgent(userAgent string) *Client {
	return defaultClient.SetUserAgent(userAgent)
}

// SetCommonBearerAuthToken is a global wrapper methods which delegated
// to the default client's SetCommonBearerAuthToken.
func SetCommonBearerAuthToken(token string) *Client {
	return defaultClient.SetCommonBearerAuthToken(token)
}

// SetCommonBasicAuth is a global wrapper methods which delegated
// to the default client's SetCommonBasicAuth.
func SetCommonBasicAuth(username, password string) *Client {
	return defaultClient.SetCommonBasicAuth(username, password)
}

// SetCommonHeaders is a global wrapper methods which delegated
// to the default client's SetCommonHeaders.
func SetCommonHeaders(hdrs map[string]string) *Client {
	return defaultClient.SetCommonHeaders(hdrs)
}

// SetCommonHeader is a global wrapper methods which delegated
// to the default client's SetCommonHeader.
func SetCommonHeader(key, value string) *Client {
	return defaultClient.SetCommonHeader(key, value)
}

// SetCommonContentType is a global wrapper methods which delegated
// to the default client's SetCommonContentType.
func SetCommonContentType(ct string) *Client {
	return defaultClient.SetCommonContentType(ct)
}

// DisableDumpAll is a global wrapper methods which delegated
// to the default client's DisableDumpAll.
func DisableDumpAll() *Client {
	return defaultClient.DisableDumpAll()
}

// SetCommonDumpOptions is a global wrapper methods which delegated
// to the default client's SetCommonDumpOptions.
func SetCommonDumpOptions(opt *DumpOptions) *Client {
	return defaultClient.SetCommonDumpOptions(opt)
}

// SetProxy is a global wrapper methods which delegated
// to the default client's SetProxy.
func SetProxy(proxy func(*http.Request) (*url.URL, error)) *Client {
	return defaultClient.SetProxy(proxy)
}

// OnBeforeRequest is a global wrapper methods which delegated
// to the default client's OnBeforeRequest.
func OnBeforeRequest(m RequestMiddleware) *Client {
	return defaultClient.OnBeforeRequest(m)
}

// OnAfterResponse is a global wrapper methods which delegated
// to the default client's OnAfterResponse.
func OnAfterResponse(m ResponseMiddleware) *Client {
	return defaultClient.OnAfterResponse(m)
}

// SetProxyURL is a global wrapper methods which delegated
// to the default client's SetProxyURL.
func SetProxyURL(proxyUrl string) *Client {
	return defaultClient.SetProxyURL(proxyUrl)
}

// DisableTraceAll is a global wrapper methods which delegated
// to the default client's DisableTraceAll.
func DisableTraceAll() *Client {
	return defaultClient.DisableTraceAll()
}

// EnableTraceAll is a global wrapper methods which delegated
// to the default client's EnableTraceAll.
func EnableTraceAll() *Client {
	return defaultClient.EnableTraceAll()
}

// SetCookieJar is a global wrapper methods which delegated
// to the default client's SetCookieJar.
func SetCookieJar(jar http.CookieJar) *Client {
	return defaultClient.SetCookieJar(jar)
}

// ClearCookies is a global wrapper methods which delegated
// to the default client's ClearCookies.
func ClearCookies() *Client {
	return defaultClient.ClearCookies()
}

// SetJsonMarshal is a global wrapper methods which delegated
// to the default client's SetJsonMarshal.
func SetJsonMarshal(fn func(v interface{}) ([]byte, error)) *Client {
	return defaultClient.SetJsonMarshal(fn)
}

// SetJsonUnmarshal is a global wrapper methods which delegated
// to the default client's SetJsonUnmarshal.
func SetJsonUnmarshal(fn func(data []byte, v interface{}) error) *Client {
	return defaultClient.SetJsonUnmarshal(fn)
}

// SetXmlMarshal is a global wrapper methods which delegated
// to the default client's SetXmlMarshal.
func SetXmlMarshal(fn func(v interface{}) ([]byte, error)) *Client {
	return defaultClient.SetXmlMarshal(fn)
}

// SetXmlUnmarshal is a global wrapper methods which delegated
// to the default client's SetXmlUnmarshal.
func SetXmlUnmarshal(fn func(data []byte, v interface{}) error) *Client {
	return defaultClient.SetXmlUnmarshal(fn)
}

// SetDialTLS is a global wrapper methods which delegated
// to the default client's SetDialTLS.
func SetDialTLS(fn func(ctx context.Context, network, addr string) (net.Conn, error)) *Client {
	return defaultClient.SetDialTLS(fn)
}

// SetDial is a global wrapper methods which delegated
// to the default client's SetDial.
func SetDial(fn func(ctx context.Context, network, addr string) (net.Conn, error)) *Client {
	return defaultClient.SetDial(fn)
}

// SetTLSHandshakeTimeout is a global wrapper methods which delegated
// to the default client's SetTLSHandshakeTimeout.
func SetTLSHandshakeTimeout(timeout time.Duration) *Client {
	return defaultClient.SetTLSHandshakeTimeout(timeout)
}

// EnableForceHTTP1 is a global wrapper methods which delegated
// to the default client's EnableForceHTTP1.
func EnableForceHTTP1() *Client {
	return defaultClient.EnableForceHTTP1()
}

// EnableForceHTTP2 is a global wrapper methods which delegated
// to the default client's EnableForceHTTP2.
func EnableForceHTTP2() *Client {
	return defaultClient.EnableForceHTTP2()
}

// EnableForceHTTP3 is a global wrapper methods which delegated
// to the default client's EnableForceHTTP3.
func EnableForceHTTP3() *Client {
	return defaultClient.EnableForceHTTP3()
}

// EnableHTTP3 is a global wrapper methods which delegated
// to the default client's EnableHTTP3.
func EnableHTTP3() *Client {
	return defaultClient.EnableHTTP3()
}

// DisableForceHttpVersion is a global wrapper methods which delegated
// to the default client's DisableForceHttpVersion.
func DisableForceHttpVersion() *Client {
	return defaultClient.DisableForceHttpVersion()
}

// EnableH2C is a global wrapper methods which delegated
// to the default client's EnableH2C.
func EnableH2C() *Client {
	return defaultClient.EnableH2C()
}

// DisableH2C is a global wrapper methods which delegated
// to the default client's DisableH2C.
func DisableH2C() *Client {
	return defaultClient.DisableH2C()
}

// DisableAllowGetMethodPayload is a global wrapper methods which delegated
// to the default client's DisableAllowGetMethodPayload.
func DisableAllowGetMethodPayload() *Client {
	return defaultClient.DisableAllowGetMethodPayload()
}

// EnableAllowGetMethodPayload is a global wrapper methods which delegated
// to the default client's EnableAllowGetMethodPayload.
func EnableAllowGetMethodPayload() *Client {
	return defaultClient.EnableAllowGetMethodPayload()
}

// SetCommonRetryCount is a global wrapper methods which delegated
// to the default client, create a request and SetCommonRetryCount for request.
func SetCommonRetryCount(count int) *Client {
	return defaultClient.SetCommonRetryCount(count)
}

// SetCommonRetryInterval is a global wrapper methods which delegated
// to the default client, create a request and SetCommonRetryInterval for request.
func SetCommonRetryInterval(getRetryIntervalFunc GetRetryIntervalFunc) *Client {
	return defaultClient.SetCommonRetryInterval(getRetryIntervalFunc)
}

// SetCommonRetryFixedInterval is a global wrapper methods which delegated
// to the default client, create a request and SetCommonRetryFixedInterval for request.
func SetCommonRetryFixedInterval(interval time.Duration) *Client {
	return defaultClient.SetCommonRetryFixedInterval(interval)
}

// SetCommonRetryBackoffInterval is a global wrapper methods which delegated
// to the default client, create a request and SetCommonRetryBackoffInterval for request.
func SetCommonRetryBackoffInterval(min, max time.Duration) *Client {
	return defaultClient.SetCommonRetryBackoffInterval(min, max)
}

// SetCommonRetryHook is a global wrapper methods which delegated
// to the default client, create a request and SetRetryHook for request.
func SetCommonRetryHook(hook RetryHookFunc) *Client {
	return defaultClient.SetCommonRetryHook(hook)
}

// AddCommonRetryHook is a global wrapper methods which delegated
// to the default client, create a request and AddCommonRetryHook for request.
func AddCommonRetryHook(hook RetryHookFunc) *Client {
	return defaultClient.AddCommonRetryHook(hook)
}

// SetCommonRetryCondition is a global wrapper methods which delegated
// to the default client, create a request and SetCommonRetryCondition for request.
func SetCommonRetryCondition(condition RetryConditionFunc) *Client {
	return defaultClient.SetCommonRetryCondition(condition)
}

// AddCommonRetryCondition is a global wrapper methods which delegated
// to the default client, create a request and AddCommonRetryCondition for request.
func AddCommonRetryCondition(condition RetryConditionFunc) *Client {
	return defaultClient.AddCommonRetryCondition(condition)
}

// SetResponseBodyTransformer is a global wrapper methods which delegated
// to the default client, create a request and SetResponseBodyTransformer for request.
func SetResponseBodyTransformer(fn func(rawBody []byte, req *Request, resp *Response) (transformedBody []byte, err error)) *Client {
	return defaultClient.SetResponseBodyTransformer(fn)
}

// SetUnixSocket is a global wrapper methods which delegated
// to the default client, create a request and SetUnixSocket for request.
func SetUnixSocket(file string) *Client {
	return defaultClient.SetUnixSocket(file)
}

// GetClient is a global wrapper methods which delegated
// to the default client's GetClient.
func GetClient() *http.Client {
	return defaultClient.GetClient()
}

// NewRequest is a global wrapper methods which delegated
// to the default client's NewRequest.
func NewRequest() *Request {
	return defaultClient.R()
}

// R is a global wrapper methods which delegated
// to the default client's R().
func R() *Request {
	return defaultClient.R()
}
