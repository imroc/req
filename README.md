# req

[![GoDoc](https://pkg.go.dev/badge/github.com/imroc/req.svg)](https://pkg.go.dev/github.com/imroc/req)

Simplified golang http client library with magic, happy sending requests, less code and more efficiency.

**Table of Contents**

* [Features](#Features)
* [Quick Start](#Quick-Start)
* [Debugging](#Debugging)
* [Path Parameter and Query Parameter](#Param)
* [Header and Cookie](#Header-Cookie)
* [Custom Client and Root Certificates](#Cert)
* [Basic Auth and Bearer Token](#Auth)
* [Download and Upload](#Download-Upload)
* [Auto-Decoding](#AutoDecode)
  
## <a name="Features">Features</a>

* Simple and chainable methods for both client-level and request-level settings, and the request-level setting takes precedence if both are set.
* Powerful and convenient debug utilites, including debug logs, performance traces, dump complete request and response content, even provide global wrapper methods to test with minimal code (see [Debugging](#Debugging).
* Detect the charset of response body and decode it to utf-8 automatically to avoid garbled characters by default (see [Auto-Decoding](#AutoDecode)).
* Exportable `Transport`, easy to integrate with existing `http.Client`, debug APIs with minimal code change.
* Easy [Download and Upload](#Download-Upload).
* Easy set header, cookie, path parameter, query parameter, form data, basic auth, bearer token, timeout, proxy, certs, redirect policy and so on for requests or clients.

## <a name="Quick-Start">Quick Start</a>

**Install**

``` sh
go get github.com/imroc/req/v2@v2.0.0-beta.1
```

**Import**

```go
import "github.com/imroc/req/v2"
```

```go
// For test, you can create and send a request with the global default
// client, use DevMode to see all details, try and suprise :)
req.DevMode()
req.Get("https://api.github.com/users/imroc")

// Create and send a request with the custom client and settings
client := req.C(). // Use C() to create a client
    SetUserAgent("my-custom-client"). // Chainable client settings
    SetTimeout(5 * time.Second).
    DevMode()
resp, err := client.R(). // Use R() to create a request
    SetHeader("Accept", "application/vnd.github.v3+json"). // Chainable request settings
    SetPathParam("username", "imroc").
    SetQueryParam("page", "1").
    SetResult(&result).
    Get("https://api.github.com/users/{username}/repos")
```

## <a name="Debugging">Debugging</a>

**Dump the Content**

```go
// Set EnableDump to true, dump all content to stdout by default,
// including both the header and body of all request and response
client := req.C().EnableDump(true)
client.R().Get("https://httpbin.org/get")

/* Output
:authority: httpbin.org
:method: GET
:path: /get
:scheme: https
user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36
accept-encoding: gzip

:status: 200
date: Wed, 26 Jan 2022 06:39:20 GMT
content-type: application/json
content-length: 372
server: gunicorn/19.9.0
access-control-allow-origin: *
access-control-allow-credentials: true

{
  "args": {},
  "headers": {
    "Accept-Encoding": "gzip",
    "Host": "httpbin.org",
    "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36",
    "X-Amzn-Trace-Id": "Root=1-61f0ec98-5958c02662de26e458b7672b"
  },
  "origin": "103.7.29.30",
  "url": "https://httpbin.org/get"
}
*/
	
// Customize dump settings with predefined convenience settings. 
client.EnableDumpOnlyHeader(). // Only dump the header of request and response
    EnableDumpAsync(). // Dump asynchronously to improve performance
    EnableDumpToFile("reqdump.log") // Dump to file without printing it out
// Send request to see the content that have been dumpped	
client.R().Get(url) 

// Enable dump with fully customized settings
opt := &req.DumpOptions{
            Output:         os.Stdout,
            RequestHeader:  true,
            ResponseBody:   true,
            RequestBody:    false,
            ResponseHeader: false,
            Async:          false,
        }
client.SetDumpOptions(opt).EnableDump(true)
client.R().Get("https://www.baidu.com/")

// Change settings dynamiclly
opt.ResponseBody = false
client.R().Get("https://www.baidu.com/")
```

**EnableDebugLog for Deeper Insights**

```go
// Logging is enabled by default, but only output the warning and error message.
// set `EnableDebugLog` to true to enable debug level logging.
client := req.C().EnableDebugLog(true)
client.R().Get("http://baidu.com/s?wd=req")
/* Output
2022/01/26 15:46:29.279368 DEBUG [req] GET http://baidu.com/s?wd=req
2022/01/26 15:46:29.469653 DEBUG [req] charset iso-8859-1 detected in Content-Type, auto-decode to utf-8
2022/01/26 15:46:29.469713 DEBUG [req] <redirect> GET http://www.baidu.com/s?wd=req
...
*/

// SetLogger with nil to disable all log
client.SetLogger(nil)

// Or customize the logger with your own implementation.
client.SetLogger(logger)
```

**EnableTrace to Analyze Performance**

```go
// Enable trace at request level
client := req.C()
resp, err := client.R().EnableTrace(true).Get("https://api.github.com/users/imroc")
if err != nil {
	log.Fatal(err)
}
ti := resp.TraceInfo() // Use `resp.Request.TraceInfo()` to avoid unnecessary copy in production
fmt.Println(ti)
fmt.Println("--------")
k, v := ti.MaxTime()
fmt.Printf("Max time is %s which tooks %v\n", k, v)

/* Output
TotalTime         : 1.342805875s
DNSLookupTime     : 7.549292ms
TCPConnectTime    : 567.833µs
TLSHandshakeTime  : 536.604041ms
FirstResponseTime : 797.466708ms
ResponseTime      : 374.875µs
IsConnReused:     : false
RemoteAddr        : 192.30.255.117:443
--------
Max time is FirstResponseTime which tooks 797.466708ms
*/

// Enable trace at client level
client.EnableTraceAll()
resp, err = client.R().Get(url)
// ...
```

**DevMode**

If you want to enable all debug features (dump, debug log and tracing), just call `DevMode()`:

```go
client := req.C().DevMode()
client.R().Get("https://imroc.cc")
```

**Test with Global Wrapper Methods**

`req` wrap methods of both `Client` and `Request` with global methods, which is delegated to default client, it's very convenient when making API test.

```go
// Call the global methods just like the Client's methods,
// so you can treat package name `req` as a Client, and
// you don't need to create any client explicitly.
req.SetTimeout(5 * time.Second).
	SetCommonBasicAuth("imroc", "123456").
	SetUserAgent("my api client").
	DevMode()

// Call the global method just like the Request's method,
// which will create request automatically using the default
// client, so you can treat package name `req` as a Request,
// and you don't need to create request explicitly.
req.SetQueryParam("page", "2").
	SetHeader("Accept", "application/json").
	Get("https://api.example.com/repos")
```

## <a name="Param">Path Parameter and Query Parameter</a>

**Set Path Parameter**

Use `SetPathParam` or `SetPathParams` to replace variable in the url path:

```go
client := req.C().DevMode()

client.R().
    SetPathParam("owner", "imroc"). // Set a path param, which will replace the vairable in url path
    SetPathParams(map[string]string{ // Set multiple path params at once 
        "repo": "req",
        "path": "README.md",
    }).Get("https://api.github.com/repos/{owner}/{repo}/contents/{path}") // path parameter will replace path variable in the url
/* Output
2022/01/23 14:43:59.114592 DEBUG [req] GET https://api.github.com/repos/imroc/req/contents/README.md
...
*/

// You can also set the common PathParam for every request on client
client.SetPathParam(k1, v1).SetPathParams(pathParams)
	
resp1, err := client.Get(url1)
...

resp2, err := client.Get(url2)
...
```

**Set Query Parameter**

Use `SetQueryParam`, `SetQueryParams` or `SetQueryString` to append url query parameter:

```go
client := req.C().DevMode()

client.R().
    SetQueryParam("a", "a"). // Set a query param, which will be encoded as query parameter in url
    SetQueryParams(map[string]string{ // Set multiple query params at once 
        "b": "b",
        "c": "c",
    }).SetQueryString("d=d&e=e"). // Set query params as a raw query string
    Get("https://api.github.com/repos/imroc/req/contents/README.md?x=x")
/* Output
2022/01/23 14:43:59.114592 DEBUG [req] GET https://api.github.com/repos/imroc/req/contents/README.md?x=x&a=a&b=b&c=c&d=d&e=e
...
*/

// You can also set the common QueryParam for every request on client
client.SetQueryParam(k, v).
    SetQueryParams(queryParams).
    SetQueryString(queryString).
	
resp1, err := client.Get(url1)
...
resp2, err := client.Get(url2)
...
```

## <a name="Header-Cookie">Header and Cookie</a>

**Set Header**
```go
// Let's dump the header to see what's going on
client := req.C().EnableDumpOnlyHeader() 

// Send a request with multiple headers and cookies
client.R().
    SetHeader("Accept", "application/json"). // Set one header
    SetHeaders(map[string]string{ // Set multiple headers at once 
        "My-Custom-Header": "My Custom Value",
        "User":             "imroc",
    }).Get("https://www.baidu.com/")

/* Output
GET / HTTP/1.1
Host: www.baidu.com
User-Agent: req/v2 (https://github.com/imroc/req)
Accept: application/json
My-Custom-Header: My Custom Value
User: imroc
Accept-Encoding: gzip

...
*/

// You can also set the common header and cookie for every request on client.
client.SetHeader(header).SetHeaders(headers)

resp1, err := client.R().Get(url1)
...
resp2, err := client.R().Get(url2)
...
```

**Set Cookie**

```go
// Let's dump the header to see what's going on
client := req.C().EnableDumpOnlyHeader() 

// Send a request with multiple headers and cookies
client.R().
    SetCookie(&http.Cookie{ // Set one cookie
        Name:     "imroc/req",
        Value:    "This is my custome cookie value",
        Path:     "/",
        Domain:   "baidu.com",
        MaxAge:   36000,
        HttpOnly: false,
        Secure:   true,
    }).SetCookies([]*http.Cookie{ // Set multiple cookies at once 
        &http.Cookie{
            Name:     "testcookie1",
            Value:    "testcookie1 value",
            Path:     "/",
            Domain:   "baidu.com",
            MaxAge:   36000,
            HttpOnly: false,
            Secure:   true,
        },
        &http.Cookie{
            Name:     "testcookie2",
            Value:    "testcookie2 value",
            Path:     "/",
            Domain:   "baidu.com",
            MaxAge:   36000,
            HttpOnly: false,
            Secure:   true,
        },
    }).Get("https://www.baidu.com/")

/* Output
GET / HTTP/1.1
Host: www.baidu.com
User-Agent: req/v2 (https://github.com/imroc/req)
Accept: application/json
Cookie: imroc/req="This is my custome cookie value"; testcookie1="testcookie1 value"; testcookie2="testcookie2 value"
Accept-Encoding: gzip

...
*/

// You can also set the common cookie for every request on client.
client.SetCookie(cookie).SetCookies(cookies)

resp1, err := client.R().Get(url1)
...
resp2, err := client.R().Get(url2)
```

## <a name="Cert">Custom Client and Root Certificates</a>

```go
client := req.R()

// Set root cert and client cert from file path
client.SetRootCertFromFile("/path/to/root/certs/pemFile1.pem", "/path/to/root/certs/pemFile2.pem", "/path/to/root/certs/pemFile3.pem"). // Set root cert from one or more pem files
    SetCertFromFile("/path/to/client/certs/client.pem", "/path/to/client/certs/client.key") // Set client cert and key cert file
	
// You can also set root cert from string
client.SetRootCertFromString("-----BEGIN CERTIFICATE-----XXXXXX-----END CERTIFICATE-----")

// And set client cert with 
cert1, err := tls.LoadX509KeyPair("/path/to/client/certs/client.pem", "/path/to/client/certs/client.key")
if err != nil {
    log.Fatalf("ERROR client certificate: %s", err)
}
// ...

// you can add more certs if you want
client.SetCert(cert1, cert2, cert3) 
```

## <a name="Auth">Basic Auth and Bearer Token</a>

```go
client := req.C()

// Set basic auth for all request
client.SetCommonBasicAuth("imroc", "123456")

// Set bearer token for all request
client.SetCommonBearerToken("MDc0ZTg5YmU4Yzc5MjAzZGJjM2ZiMzkz")

// Set basic auth for a request, will override client's basic auth setting.
client.R().SetBasicAuth("myusername", "mypassword").Get("https://api.example.com/profile")

// Set bearer token for a request, will override client's bearer token setting.
client.R().SetBearerToken("NGU1ZWYwZDJhNmZhZmJhODhmMjQ3ZDc4").Get("https://api.example.com/profile")
```

## <a name="Download-Upload">Download and Upload</a>

**Download**

```go
// Create a client with default download direcotry
client := req.C().SetOutputDirectory("/path/to/download").EnableDumpNoResponseBody()

// Download to relative file path, this will be downloaded
// to /path/to/download/test.jpg
client.R().SetOutputFile("test.jpg").Get(url)

// Download to absolute file path, ignore the output directory
// setting from Client
client.R().SetOutputFile("/tmp/test.jpg").Get(url)

// You can also save file to any `io.WriteCloser`
file, err := os.Create("/tmp/test.jpg")
if err != nil {
	fmt.Println(err)
	return
}
client.R().SetOutput(file).Get(url)
```

**Multipart Upload**

```go
client := req.().EnableDumpNoRequestBody() // Request body contains unreadable binary, do not dump

client.R().SetFile("pic", "test.jpg"). // Set form param name and filename
    SetFile("pic", "/path/to/roc.png"). // Multiple files using the same form param name
    SetFiles(map[string]string{ // Set multiple files using map
        "exe": "test.exe",
        "src": "main.go",
    }).
    SetFormData(map[string]string{ // Set from param using map
        "name":  "imroc",
        "email": "roc@imroc.cc",
    }).
	SetFromDataFromValues(values). // You can also set form data using `url.Values`
    Post("http://127.0.0.1:8888/upload")
*/

```

## <a name="AutoDecode">Auto-Decoding</a>

`Req` detect the charset of response body and decode it to utf-8 automatically to avoid garbled characters by default.

Its principle is to detect whether `Content-Type` header at first, if it's not the text content type (json, xml, html and so on), `req` will not try to decode. If it is, then `req` will try to find the charset information, if it's not included in the header, it will try to sniff the body's content to determine the charset, if found and is not utf-8, then decode it to utf-8 automatically, if the charset is not sure, it will not decode, and leave the body untouched.

You can also disable if you don't need or care a lot about performance:

```go
client.DisableAutoDecode(true)
```

Also you can make some customization:

```go
// Try to auto-detect and decode all content types (some server may return incorrect Content-Type header)
client.SetAutoDecodeAllType()

// Only auto-detect and decode content which `Content-Type` header contains "html" or "json"
client.SetAutoDecodeContentType("html", "json")

// Or you can customize the function to determine whether to decode
fn := func(contentType string) bool {
    if regexContentType.MatchString(contentType) {
        return true
    }
    return false
}
client.SetAutoDecodeAllTypeFunc(fn)
```

## License

`Req` released under MIT license, refer [LICENSE](LICENSE) file.