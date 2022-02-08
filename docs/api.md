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
  * [URL Query and Path Parameter](#Query)
  * [Header and Cookie](#Header)
  * [Body and Marshal&Unmarshal](#Body)
  * [Request Level Debug](#Debug-Request)
  * [Multipart & Form & Upload](#Multipart)
  * [Download](#Download)
  * [Other Settings](#Other-Request)

## <a name="Client">Client Settings</a>

The following are the chainable settings of Client, all of which have corresponding global wrappers.

Basically, you can know the meaning of most settings directly from the method name.

### <a name="Debug">Debug Features</a>

* [DevMode()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DevMode) - Enable all debug features (Dump, DebugLog and Trace).

* [EnableDebugLog()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDebugLog) - Enable debug level log (disabled by default).
* [DisableDebugLog()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableDebugLog)
* [SetLogger(log Logger)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetLogger) - Set the customized logger.

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
* [SetCommonDumpOptions(opt *DumpOptions)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonDumpOptions)

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
* [DisableAutoDecode()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableAutoDecode) - Disable auto-detect charset and decode to utf-8 (enabled by default).
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

The following are the chainable settings of Request, all of which have corresponding global wrappers.

Basically, you can know the meaning of most settings directly from the method name.

### <a name="Query">URL Query and Path Parameter</a>

* [AddQueryParam(key, value string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.AddQueryParam)
* [SetQueryParam(key, value string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetQueryParam)
* [SetQueryParams(params map[string]string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetQueryParams)
* [SetQueryString(query string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetQueryString)
* [SetPathParam(key, value string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetPathParam)
* [SetPathParams(params map[string]string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetPathParams)

### <a name="Header">Header and Cookie</a>

* [SetHeader(key, value string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetHeader)
* [SetHeaders(hdrs map[string]string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetHeaders)
* [SetBasicAuth(username, password string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetBasicAuth)
* [SetBearerAuthToken(token string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetBearerAuthToken)
* [SetContentType(contentType string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetContentType)
* [SetCookies(cookies ...*http.Cookie)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetCookies)

### <a name="Body">Body and Marshal&Unmarshal</a>

* [SetBody(body interface{})](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetBody)
* [SetBodyBytes(body []byte)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetBodyBytes)
* [SetBodyJsonBytes(body []byte)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetBodyJsonBytes)
* [SetBodyJsonMarshal(v interface{})](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetBodyJsonMarshal)
* [SetBodyJsonString(body string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetBodyJsonString)
* [SetBodyString(body string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetBodyString)
* [SetBodyXmlBytes(body []byte)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetBodyXmlBytes)
* [SetBodyXmlMarshal(v interface{})](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetBodyXmlMarshal)
* [SetBodyXmlString(body string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetBodyXmlString)
* [SetResult(result interface{})](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetResult)
* [SetError(error interface{})](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetError)

### <a name="Debug-Request">Request Level Debug</a>

* [EnableTrace()](https://pkg.go.dev/github.com/imroc/req/v3#Request.EnableTrace) - Disabled by default.
* [DisableTrace()](https://pkg.go.dev/github.com/imroc/req/v3#Request.DisableTrace)
* [EnableDump()](https://pkg.go.dev/github.com/imroc/req/v3#Request.EnableDump)
* [EnableDumpTo(output io.Writer)](https://pkg.go.dev/github.com/imroc/req/v3#Request.EnableDumpTo)
* [EnableDumpToFile(filename string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.EnableDumpToFile)
* [EnableDumpWithoutBody()](https://pkg.go.dev/github.com/imroc/req/v3#Request.EnableDumpWithoutBody)
* [EnableDumpWithoutHeader()](https://pkg.go.dev/github.com/imroc/req/v3#Request.EnableDumpWithoutHeader)
* [EnableDumpWithoutRequest()](https://pkg.go.dev/github.com/imroc/req/v3#Request.EnableDumpWithoutRequest)
* [EnableDumpWithoutRequestBody()](https://pkg.go.dev/github.com/imroc/req/v3#Request.EnableDumpWithoutRequestBody)
* [EnableDumpWithoutResponse()](https://pkg.go.dev/github.com/imroc/req/v3#Request.EnableDumpWithoutResponse)
* [EnableDumpWithoutResponseBody()](https://pkg.go.dev/github.com/imroc/req/v3#Request.EnableDumpWithoutResponseBody)
* [SetDumpOptions(opt *DumpOptions)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetDumpOptions)

### <a name="Multipart">Multipart & Form & Upload</a>

* [SetFormData(data map[string]string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetFormData)
* [SetFormDataFromValues(data url.Values)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetFormDataFromValues)
* [SetFile(paramName, filePath string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetFile)
* [SetFileReader(paramName, filePath string, reader io.Reader)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetFileReader)
* [SetFiles(files map[string]string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetFiles)

### <a name="Download">Download</a>

* [SetOutput(output io.Writer)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetOutput)
* [SetOutputFile(file string)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetOutputFile)

### <a name="Other-Request">Other Settings</a>

* [SetContext(ctx context.Context)](https://pkg.go.dev/github.com/imroc/req/v3#Request.SetContext)
