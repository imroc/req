<p align="center">
    <h1 align="center">Req API Reference </h1>
</p>

Here is a brief list of some core APIs, which is convenient to get started quickly. For a more detailed and complete list of APIs, please refer to [GoDoc](https://pkg.go.dev/github.com/imroc/req/v3).

## Table of Contents

* [Client Settings](#Client)
  * [Debug Features](#Debug)
  * [Common Settings for HTTP Requests](#Common)
  * [Auto-Decode](#Decode)
  * [Certificates](#Certs)
  * [Marshal&Unmarshal](#Marshal)
  * [Other Settings](#Other)
* [Request Settings](#Request)

## <a name="Client">Client Settings</a>

The following are the chainable settings Client, all of which have corresponding global wrappers.

### <a name="Debug">Debug Features</a>

* [DevMode()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DevMode) - Enable all debug features (Dump, DebugLog and Trace).

* [EnableDebugLog()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDebugLog) - Enable debug level log (disabled by default).
* [DisableDebugLog()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableDebugLog)
* [SetLogger(log Logger)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetLogger) - Set the customized logger, set to nil to disable logger.

* [EnableDumpAll()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAll) - Enable dump for all requests.
* [DisableDumpAll()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableDumpAll)
* [EnableDumpAllWithoutResponseBody()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllWithoutResponseBody)
* [EnableDumpAllWithoutResponse()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllWithoutResponse)
* [EnableDumpAllWithoutRequestBody()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllWithoutRequestBody)
* [EnableDumpAllWithoutRequest()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllWithoutRequest)
* [EnableDumpAllWithoutHeader()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllWithoutHeader)
* [EnableDumpAllWithoutBody()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllWithoutBody)
* [EnableDumpAllToFile(filename string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllToFile)
* [EnableDumpAllTo(output io.Writer)](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllTo)
* [EnableDumpAllAsync()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllAsync)
* [SetCommonDumpOptions(opt *DumpOptions)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonDumpOptions) - need to call `EnableDumpAll()` if you want to enable dump.

* [EnableTraceAll()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableTraceAll) - Enable trace for all requests (disabled by default).
* [DisableTraceAll()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableTraceAll)

### <a name="Common">Common Settings for HTTP Requests</a>

* [SetCommonQueryString(query string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonQueryString)
* [SetCommonHeaders(hdrs map[string]string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonHeaders)
* [SetCommonHeader(key, value string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonHeader)
* [SetCommonContentType(ct string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonContentType)
* [SetCommonBearerAuthToken(token string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonBearerAuthToken)
* [SetCommonBasicAuth(username, password string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonBasicAuth)
* [SetCommonCookies(cookies ...*http.Cookie)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonCookies)
* [AddCommonQueryParam(key, value string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.AddCommonQueryParam)
* [SetCommonQueryParam(key, value string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonQueryParam)
* [SetCommonQueryParams(params map[string]string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonQueryParams)
* [SetCommonFormDataFromValues(data url.Values)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonFormDataFromValues)
* [SetCommonFormData(data map[string]string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonFormData)
* [SetUserAgent(userAgent string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetUserAgent)

### <a name="Decode">Auto-Decode</a>

* [EnableAutoDecode()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableAutoDecode)
* [DisableAutoDecode()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableAutoDecode) - Disable auto-detect charset and decode to utf-8 (enabled by default)
* [SetAutoDecodeContentType(contentTypes ...string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetAutoDecodeContentType)
* [SetAutoDecodeAllContentType()](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetAutoDecodeAllContentType)
* [SetAutoDecodeContentTypeFunc(fn func(contentType string) bool)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetAutoDecodeContentTypeFunc)

### <a name="Certs">Certificates</a>

* [SetCerts(certs ...tls.Certificate) ](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCerts)
* [SetCertFromFile(certFile, keyFile string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCertFromFile)
* [SetRootCertsFromFile(pemFiles ...string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetRootCertsFromFile)
* [SetRootCertFromString(pemContent string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetRootCertFromString)

### <a name="Marshal">Marshal&Unmarshal</a>

* [SetJsonUnmarshal(fn func(data []byte, v interface{}) error)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetJsonUnmarshal)
* [SetJsonMarshal(fn func(v interface{}) ([]byte, error))](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetJsonMarshal)
* [SetXmlMarshal(fn func(v interface{}) ([]byte, error))](https://pkg.go.dev/github.com/imroc/req/v3#SetXmlUnmarshal)
* [SetXmlMarshal(fn func(v interface{}) ([]byte, error))](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetXmlMarshal)

### <a name="Middleware">Middleware</a>

* [OnBeforeRequest(m RequestMiddleware)](https://pkg.go.dev/github.com/imroc/req/v3#Client.OnBeforeRequest)
* [OnAfterResponse(m ResponseMiddleware)](https://pkg.go.dev/github.com/imroc/req/v3#Client.OnAfterResponse)

### <a name="Other">Other Settings</a>

* [SetTimeout(d time.Duration)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetTimeout)
* [SetTLSHandshakeTimeout(timeout time.Duration)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetTLSHandshakeTimeout)
* [SetTLSClientConfig(conf *tls.Config)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetTLSClientConfig)

* [EnableForceHTTP1()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableForceHTTP1) - Disabled by default.
* [DisableForceHTTP1()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableForceHTTP1)

* [EnableKeepAlives()](EnableKeepAlives())
* [DisableKeepAlives()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableKeepAlives) Enabled by default.

* [SetScheme(scheme string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetScheme)
* [SetBaseURL(u string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetBaseURL)

* [SetProxyURL(proxyUrl string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetProxyURL)
* [SetProxy(proxy func(*http.Request) (*urlpkg.URL, error))](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetProxy)

* [SetOutputDirectory(dir string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetOutputDirectory)

* [SetDialTLS(fn func(ctx context.Context, network, addr string) (net.Conn, error))](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetDialTLS)
* [SetDial(fn func(ctx context.Context, network, addr string) (net.Conn, error))](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetDial)

* [SetCookieJar(jar http.CookieJar)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCookieJar)

* [SetRedirectPolicy(policies ...RedirectPolicy)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetRedirectPolicy)

* [EnableCompression()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableCompression)
* [DisableCompression()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableCompression) - Enabled by default

* [EnableAutoReadResponse()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableAutoReadResponse)
* [DisableAutoReadResponse()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableAutoReadResponse) - Enabled by default

* [EnableAllowGetMethodPayload()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableAllowGetMethodPayload) - Disabled by default. 
* [DisableAllowGetMethodPayload()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableAllowGetMethodPayload)

## <a name="Request">Request Settings</a>

The following are the chainable settings Request, all of which have corresponding global wrappers.

[AddQueryParam(key, value string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.AddQueryParam)
[DisableTrace()](https://pkg.go.dev/github.com/imroc/req/v3#Request.DisableTrace)
[EnableDump()](https://pkg.go.dev/github.com/imroc/req/v3#Request.EnableDump)