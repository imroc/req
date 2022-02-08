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

## <a name="Client">Client Settings</a>

The following are the chainable settings Client, all of which have corresponding global wrappers.

### <a name="Debug">Debug Features</a>

* [DevMode()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DevMode) - Enable all debug features (Dump, DebugLog and Trace).

* [EnableDebugLog()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDebugLog) - Enable debug level log (disabled by default).
* [DisableDebugLog()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableDebugLog) - Disable debug level log (disabled by default).
* [SetLogger(log Logger)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetLogger) - Set the customized logger, set to nil to disable logger.

* [EnableDumpAll()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAll) - Enable dump for all requests, including all content for the request and response by default.
* [DisableDumpAll()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableDumpAll) - Disable dump for all requests.
* [EnableDumpAllWithoutResponseBody()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllWithoutResponseBody) - Enable dump for all requests without response body, can be used in download request to avoid dump the unreadable binary content.
* [EnableDumpAllWithoutResponse()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllWithoutResponse) - Enable dump for all requests without response (only request header and body).
* [EnableDumpAllWithoutRequestBody()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllWithoutRequestBody) - Enable dump for all requests without request body, can be used in upload request to avoid dump the unreadable binary content.
* [EnableDumpAllWithoutRequest()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllWithoutRequest) - Enable dump for all requests without request (only response header and body).
* [EnableDumpAllWithoutHeader()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllWithoutHeader) - Enable dump for all requests without header (only body of request and response).
* [EnableDumpAllWithoutBody()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllWithoutBody) - Enable dump for all requests without body (only header of request and response).
* [EnableDumpAllToFile(filename string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllToFile) - Enable dump for all requests and save to the specified filename.
* [EnableDumpAllTo(output io.Writer)](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllTo) - Enables dump for all requests and save to the specified `io.Writer`.
* [EnableDumpAllAsync()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableDumpAllAsync) - Enable dump for all requests and output asynchronously, can be used for debugging in production environment without affecting performance.
* [SetCommonDumpOptions(opt *DumpOptions)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonDumpOptions) -  Configures the underlying Transport's `DumpOptions` (need to call `EnableDumpAll()` if you want to enable dump).

* [EnableTraceAll()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableTraceAll) - Enable trace for all requests (disabled by default).
* [DisableTraceAll()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableTraceAll) - Disable trace at client level (disabled by default).

### <a name="Common">Common Settings for HTTP Requests</a>

* [SetCommonQueryString(query string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonQueryString) - Set a URL query parameters for all requests using the raw query string.
* [SetCommonHeaders(hdrs map[string]string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonHeaders) - Set headers for all requests from a map.
* [SetCommonHeader(key, value string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonHeader) - Set a header with key-value pair for all requests.
* [SetCommonContentType(ct string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonContentType) - Set the `Content-Type` header for all requests.
* [SetCommonBearerAuthToken(token string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonBearerAuthToken) - Set the bearer auth token for all requests.
* [SetCommonBasicAuth(username, password string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonBasicAuth) - Set the basic auth for all requests.
* [SetCommonCookies(cookies ...*http.Cookie)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonCookies) - Set cookies for all requests.
* [AddCommonQueryParam(key, value string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.AddCommonQueryParam) - Add a URL query parameter with key-value pair for all requests which will not override if same key exists.
* [SetCommonQueryParam(key, value string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonQueryParam) - Set a URL query parameter with key-value pair for all requests.
* [SetCommonQueryParams(params map[string]string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonQueryParams) - Set the URL query parameters with a map for all requests.
* [SetCommonFormDataFromValues(data url.Values)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonFormDataFromValues) - Set the form data from `url.Values` for all requests.
* [SetCommonFormData(data map[string]string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCommonFormData) - Set the form data from map for all requests.
* [SetUserAgent(userAgent string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetUserAgent) - Set the "User-Agent" header for all requests.

### <a name="Decode">Auto-Decode</a>

* [EnableAutoDecode()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableAutoDecode) - Enable auto-detect charset and decode to utf-8 (enabled by default).
* [DisableAutoDecode()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableAutoDecode) - Disable auto-detect charset and decode to utf-8 (enabled by default)
* [SetAutoDecodeContentType(contentTypes ...string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetAutoDecodeContentType) - Set the content types that will be auto-detected and decode to utf-8.
* [SetAutoDecodeAllContentType()](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetAutoDecodeAllContentType) - Set try to auto-detect and decode all content type to utf-8.
* [SetAutoDecodeContentTypeFunc(fn func(contentType string) bool)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetAutoDecodeContentTypeFunc) - Custmize the function that determines the content type whether it should be auto-detected and decode to utf-8.

### <a name="Certs">Certificates</a>

* [SetCerts(certs ...tls.Certificate) ](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCerts) - Set client certificates from on one more `tls.Certificate`.
* [SetCertFromFile(certFile, keyFile string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCertFromFile) - Set client certificates from cert and key file.
* [SetRootCertsFromFile(pemFiles ...string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetRootCertsFromFile) - Set root certificates from pem files.
* [SetRootCertFromString(pemContent string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetRootCertFromString) - Set root certificates from string.

### <a name="Marshal">Marshal&Unmarshal</a>

* [SetJsonUnmarshal(fn func(data []byte, v interface{}) error)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetJsonUnmarshal) - Set the JSON Unmarshal function which will be used to unmarshal response body.
* [SetJsonMarshal(fn func(v interface{}) ([]byte, error))](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetJsonMarshal) - Set JSON Marshal function which will be used to marshal request body.
* [SetXmlMarshal(fn func(v interface{}) ([]byte, error))](https://pkg.go.dev/github.com/imroc/req/v3#SetXmlUnmarshal) - Set the XML Unmarshal function which will be used to unmarshal response body.
* [SetXmlMarshal(fn func(v interface{}) ([]byte, error))](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetXmlMarshal) - Set the XML Marshal function which will be used to marshal request body.

### <a name="Other">Other Settings</a>

* [OnBeforeRequest(m RequestMiddleware)](https://pkg.go.dev/github.com/imroc/req/v3#Client.OnBeforeRequest) - Add a request middleware which hooks before request sent.
* [OnAfterResponse(m ResponseMiddleware)](https://pkg.go.dev/github.com/imroc/req/v3#Client.OnAfterResponse) - Add a response middleware which hooks after response received.

* [SetTLSClientConfig(conf *tls.Config)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetTLSClientConfig) - Set the client tls config.
* [SetTLSHandshakeTimeout(timeout time.Duration)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetTLSHandshakeTimeout) - Set the TLS handshake timeout.

* [EnableForceHTTP1()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableForceHTTP1) - Enable force using HTTP1 (disabled by default).
* [DisableForceHTTP1()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableForceHTTP1) - Disable force using HTTP1.

* [EnableKeepAlives()](EnableKeepAlives()) - Enable HTTP keep-alives.
* [DisableKeepAlives()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableKeepAlives) - Disable HTTP keep-alives (enabled by default)

* [SetTimeout(d time.Duration)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetTimeout) - Set the request timeout.

* [SetScheme(scheme string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetScheme) - Set the default scheme in the client, will be used when there is no scheme in the request url (e.g. "github.com/imroc/req").
* [SetBaseURL(u string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetBaseURL) - Set the default base url, will be used if request url is a relative url.

* [SetProxyURL(proxyUrl string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetProxyURL) - Set proxy from the proxy URL.
* [SetProxy(proxy func(*http.Request) (*urlpkg.URL, error))](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetProxy) - Set proxy from proxy function (e.g. [http.ProxyFromEnvironment](https://pkg.go.dev/net/http@go1.17.6#ProxyFromEnvironment)).

* [SetOutputDirectory(dir string)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetOutputDirectory) - Set output directory that response body will be downloaded to.

* [SetDialTLS(fn func(ctx context.Context, network, addr string) (net.Conn, error))](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetDialTLS) - Set the customized `DialTLSContext` function to Transport (make sure the returned `conn` implements [TLSConn](https://pkg.go.dev/github.com/imroc/req/v3#TLSConn) if you want your customized `conn` supports HTTP2).
* [SetDial(fn func(ctx context.Context, network, addr string) (net.Conn, error))](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetDial) - Set the customized `DialContext` function to Transport.

* [SetCookieJar(jar http.CookieJar)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetCookieJar) - Set the `CookeJar` to the underlying `http.Client`.

* [SetRedirectPolicy(policies ...RedirectPolicy)](https://pkg.go.dev/github.com/imroc/req/v3#Client.SetRedirectPolicy) - Set the RedirectPolicy (see the predefined [AllowedDomainRedirectPolicy](https://pkg.go.dev/github.com/imroc/req/v3#AllowedDomainRedirectPolicy), [AllowedHostRedirectPolicy](https://pkg.go.dev/github.com/imroc/req/v3#AllowedHostRedirectPolicy), [MaxRedirectPolicy](https://pkg.go.dev/github.com/imroc/req/v3#MaxRedirectPolicy), [NoRedirectPolicy](https://pkg.go.dev/github.com/imroc/req/v3#NoRedirectPolicy), [SameDomainRedirectPolicy](https://pkg.go.dev/github.com/imroc/req/v3#SameDomainRedirectPolicy), [SameHostRedirectPolicy](https://pkg.go.dev/github.com/imroc/req/v3#SameDomainRedirectPolicy)).

* [EnableCompression()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableCompression) - Enable the compression (enabled by default).
* [DisableCompression()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableCompression) - Disable the compression.

* [EnableAutoReadResponse()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableAutoReadResponse) - Enable read response body automatically (enabled by default).
* [DisableAutoReadResponse()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableAutoReadResponse) - Disable read response body automatically.

* [EnableAllowGetMethodPayload()](https://pkg.go.dev/github.com/imroc/req/v3#Client.EnableAllowGetMethodPayload) - Enable allow sending GET method requests with body (disabled by default). 
* [DisableAllowGetMethodPayload()](https://pkg.go.dev/github.com/imroc/req/v3#Client.DisableAllowGetMethodPayload) - Disable allow sending GET method requests with body.