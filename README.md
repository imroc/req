<p align="center">
    <h1 align="center">Req</h1>
    <p align="center">Simplified Golang HTTP client library with Black Magic, Less Code and More Efficiency.</p>
    <p align="center"><a href="https://pkg.go.dev/github.com/imroc/req/v3"><img src="https://pkg.go.dev/badge/github.com/imroc/req.svg"></a></p>
</p>

## Big News

Brand new v3 version is out, which is completely rewritten, bringing revolutionary innovations and many superpowers, try and enjoy :)

If you want to use the older version, check it out on [v1 branch](https://github.com/imroc/req/tree/v1).

> v2 is a transitional version, cuz some breaking changes were introduced during v2 refactoring, checkout [v2 branch](https://github.com/imroc/req/tree/v2) if you want.

## Table of Contents

* [Features](#Features)
* [Quick Start](#Quick-Start)
* [Debugging - Log/Trace/Dump](#Debugging)
* [Quick HTTP Test](#Test)
* [URL Path and Query Parameter](#Param)
* [Form Data](#Form)
* [Header and Cookie](#Header-Cookie)
* [Body and Marshal/Unmarshal](#Body)
* [Custom Certificates](#Cert)
* [Basic Auth and Bearer Token](#Auth)
* [Download and Upload](#Download-Upload)
* [Auto-Decode](#AutoDecode)
* [Request and Response Middleware](#Middleware)
* [Redirect Policy](#Redirect)
* [Proxy](#Proxy)
* [TODO List](#TODO)
* [License](#License)

## <a name="Features">Features</a>

* Simple and chainable methods for both client-level and request-level settings, and the request-level setting takes precedence if both are set.
* Powerful and convenient debug utilites, including debug logs, performance traces, dump complete request and response content, and even provide global wrapper methods to test with minimal code (see [Debugging - Log/Trace/Dump](#Debugging).
* Easy making HTTP test with code instead of tools like curl or postman, `req` provide global wrapper methods and `MustXXX` to test API with minimal code (see [Quick HTTP Test](#Test)).
* Detect the charset of response body and decode it to utf-8 automatically to avoid garbled characters by default (see [Auto-Decode](#AutoDecode)).
* Automatic marshal and unmarshal for JSON and XML content type and fully customizable.
* Works fine both with `HTTP/2` and `HTTP/1.1`, `HTTP/2` is preferred by default if server support.
* Exportable `Transport`, easy to integrate with existing `http.Client`, debug APIs with minimal code change.
* Easy [Download and Upload](#Download-Upload).
* Easy set header, cookie, path parameter, query parameter, form data, basic auth, bearer token for both client and request level.
* Easy set timeout, proxy, certs, redirect policy, cookie jar, compression, keepalives etc for client.
* Support middleware before request sent and after got response.

## <a name="Quick-Start">Quick Start</a>

**Install**

``` sh
go get github.com/imroc/req/v3
```

**Import**

```go
import "github.com/imroc/req/v3"
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

Checkout more runnable examples in the [examples](examples) direcotry.

## <a name="Debugging">Debugging - Log/Trace/Dump</a>

**Dump the Content**

```go
// Enable dump at client level, which will dump for all requests,
// including all content of request and response and output
// to stdout by default.
client := req.C().EnableDumpAll()
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
	
// Customize client level dump settings with predefined convenience settings. 
client.EnableDumpAllWithoutBody(). // Only dump the header of request and response
    EnableDumpAllAsync(). // Dump asynchronously to improve performance
    EnableDumpAllToFile("reqdump.log") // Dump to file without printing it out
// Send request to see the content that have been dumpped	
client.R().Get(url) 

// Enable dump with fully customized settings at client level.
opt := &req.DumpOptions{
            Output:         os.Stdout,
            RequestHeader:  true,
            ResponseBody:   true,
            RequestBody:    false,
            ResponseHeader: false,
            Async:          false,
        }
client.SetCommonDumpOptions(opt).EnableDumpAll()
client.R().Get("https://www.baidu.com/")

// Change settings dynamiclly
opt.ResponseBody = false
client.R().Get("https://www.baidu.com/")

// You can also enable dump at request level, dump to memory and will not print it out
// by default, you can call `Response.Dump()` to get the dump result and print
// only if you want to.
resp, err := client.R().EnableDump().SetBody("test body").Post("https://httpbin.org/post")
if err != nil {
    fmt.Println("err:", err)
    fmt.Println("raw content:\n", resp.Dump())
    return
}
if resp.StatusCode > 299 {
    fmt.Println("bad status:", resp.Status)
    fmt.Println("raw content:\n", resp.Dump())
	return
}

// Similarly, also support to customize dump settings with predefined convenience settings at request level.
resp, err = client.R().EnableDumpWithoutRequest().SetBody("test body").Post("https://httpbin.org/post")
// ...
resp, err = client.R().SetDumpOptions(opt).EnableDump().SetBody("test body").Post("https://httpbin.org/post")
```

**Enable DebugLog for Deeper Insights**

```go
// Logging is enabled by default, but only output the warning and error message.
// Use `EnableDebugLog` to enable debug level logging.
client := req.C().EnableDebugLog()
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

**Enable Trace to Analyze Performance**

```go
// Enable trace at request level
client := req.C()
resp, err := client.R().EnableTrace().Get("https://api.github.com/users/imroc")
if err != nil {
	log.Fatal(err)
}
trace := resp.TraceInfo() // Use `resp.Request.TraceInfo()` to avoid unnecessary struct copy in production.
fmt.Println(trace.Blame()) // Print out exactly where the http request is slowing down.
fmt.Println("----------")
fmt.Println(trace) // Print details

/* Output
the request total time is 2.562416041s, and costs 1.289082208s from connection ready to server respond frist byte
--------
TotalTime         : 2.562416041s
DNSLookupTime     : 445.246375ms
TCPConnectTime    : 428.458µs
TLSHandshakeTime  : 825.888208ms
FirstResponseTime : 1.289082208s
ResponseTime      : 1.712375ms
IsConnReused:     : false
RemoteAddr        : 98.126.155.187:443
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

## <a name="Test">Quick HTTP Test</a>

**Test with Global Wrapper Methods**

`req` wrap methods of both `Client` and `Request` with global methods, which is delegated to default client, it's very convenient when making API test.

```go
// Call the global methods just like the Client's methods,
// so you can treat package name `req` as a Client, and
// you don't need to create any client explicitly.
req.SetTimeout(5 * time.Second).
	SetCommonBasicAuth("imroc", "123456").
	SetCommonHeader("Accept", "application/json").
	SetUserAgent("my api client").
	DevMode()

// Call the global method just like the Request's method,
// which will create request automatically using the default
// client, so you can treat package name `req` as a Request,
// and you don't need to create request explicitly.
req.SetQueryParam("page", "2").
	SetHeader("Accept", "text/xml"). // Override client level settings at request level.
	Get("https://api.example.com/repos")
```

**Test with MustXXX**

Use `MustXXX` to ignore error handling during test, make it possible to complete a complex test with just one line of code:

```go
fmt.Println(req.DevMode().R().MustGet("https://imroc.cc").TraceInfo())
```

## <a name="Param">URL Path and Query Parameter</a>

**Path Parameter**

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
client.SetCommonPathParam(k1, v1).SetCommonPathParams(pathParams)
	
resp1, err := client.Get(url1)
...

resp2, err := client.Get(url2)
...
```

**Query Parameter**

Use `SetQueryParam`, `SetQueryParams` or `SetQueryString` to append url query parameter:

```go
client := req.C().DevMode()

// Set query parameter at request level.
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

// You can also set the query parameter at client level.
client.SetCommonQueryParam(k, v).
    SetCommonQueryParams(queryParams).
    SetCommonQueryString(queryString).
	
resp1, err := client.Get(url1)
...
resp2, err := client.Get(url2)
...

// Add query parameter with multiple values at request level.
client.R().AddQueryParam("key", "value1").AddQueryParam("key", "value2").Get("https://httpbin.org/get")
/* Output
2022/02/05 08:49:26.260780 DEBUG [req] GET https://httpbin.org/get?key=value1&key=value2
...
 */


// Multiple values also supported at client level.
client.AddCommonQueryParam("key", "value1").AddCommonQueryParam("key", "value2")
```

## <a name="Form">Form Data</a>

```go
client := req.C().EnableDumpOnlyRequest()
client.R().SetFormData(map[string]string{
    "username": "imroc",
    "blog":     "https://imroc.cc",
}).Post("https://httpbin.org/post")
/* Output
:authority: httpbin.org
:method: POST
:path: /post
:scheme: https
content-type: application/x-www-form-urlencoded
accept-encoding: gzip
user-agent: req/v2 (https://github.com/imroc/req)

blog=https%3A%2F%2Fimroc.cc&username=imroc
*/

// Multi value form data
v := url.Values{
    "multi": []string{"a", "b", "c"},
}
client.R().SetFormDataFromValues(v).Post("https://httpbin.org/post")
/* Output
:authority: httpbin.org
:method: POST
:path: /post
:scheme: https
content-type: application/x-www-form-urlencoded
accept-encoding: gzip
user-agent: req/v2 (https://github.com/imroc/req)

multi=a&multi=b&multi=c
*/

// You can also set form data in client level
client.SetCommonFormData(m)
client.SetCommonFormDataFromValues(v)
```

> `GET`, `HEAD`, and `OPTIONS` requests ignores form data by default

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
*/

// You can also set the common header and cookie for every request on client.
client.SetCommonHeader(header).SetCommonHeaders(headers)

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
    SetCookies(
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
    ).Get("https://www.baidu.com/")

/* Output
GET / HTTP/1.1
Host: www.baidu.com
User-Agent: req/v2 (https://github.com/imroc/req)
Accept: application/json
Cookie: testcookie1="testcookie1 value"; testcookie2="testcookie2 value"
Accept-Encoding: gzip
*/

// You can also set the common cookie for every request on client.
client.SetCommonCookies(cookie1, cookie2, cookie3)

resp1, err := client.R().Get(url1)
...
resp2, err := client.R().Get(url2)
```

You can also customize the CookieJar:
```go
// Set your own http.CookieJar implementation
client.SetCookieJar(jar)

// Set to nil to disable CookieJar
client.SetCookieJar(nil)
```

## <a name="Body">Body and Marshal/Unmarshal</a>

**Request Body**

```go
// Create a client that dump request
client := req.C().EnableDumpOnlyRequest()
// SetBody accepts string, []byte, io.Reader, use type assertion to
// determine the data type of body automatically. 
client.R().SetBody("test").Post("https://httpbin.org/post")
/* Output
:authority: httpbin.org
:method: POST
:path: /post
:scheme: https
accept-encoding: gzip
user-agent: req/v2 (https://github.com/imroc/req)

test
*/

// If it cannot determine, like map and struct, then it will wait
// and marshal to JSON or XML automatically according to the `Content-Type`
// header that have been set before or after, default to json if not set.
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
user := &User{Name: "imroc", Email: "roc@imroc.cc"}
client.R().SetBody(user).Post("https://httpbin.org/post")
/* Output
:authority: httpbin.org
:method: POST
:path: /post
:scheme: https
content-type: application/json; charset=utf-8
accept-encoding: gzip
user-agent: req/v2 (https://github.com/imroc/req)

{"name":"imroc","email":"roc@imroc.cc"}
*/


// You can use more specific methods to avoid type assertions and improves performance,
client.R().SetBodyJsonString(`{"username": "imroc"}`).Post("https://httpbin.org/post")
/*
:authority: httpbin.org
:method: POST
:path: /post
:scheme: https
content-type: application/json; charset=utf-8
accept-encoding: gzip
user-agent: req/v2 (https://github.com/imroc/req)

{"username": "imroc"}
*/

// Marshal body and set `Content-Type` automatically without any guess
cient.R().SetBodyXmlMarshal(user).Post("https://httpbin.org/post")
/* Output
:authority: httpbin.org
:method: POST
:path: /post
:scheme: https
content-type: text/xml; charset=utf-8
accept-encoding: gzip
user-agent: req/v2 (https://github.com/imroc/req)

<User><Name>imroc</Name><Email>roc@imroc.cc</Email></User>
*/
```

**Response Body**

```go
// Define success body struct
type User struct {
    Name string `json:"name"`
    Blog string `json:"blog"`
}
// Define error body struct
type ErrorMessage struct {
    Message string `json:"message"`
}
// Create a client and dump body to see details
client := req.C().EnableDumpOnlyBody()

// Send a request and unmarshal result automatically according to
// response `Content-Type`
user := &User{}
errMsg := &ErrorMessage{}
resp, err := client.R().
	SetResult(user). // Set success result
	SetError(errMsg). // Set error result
	Get("https://api.github.com/users/imroc")
if err != nil {
    log.Fatal(err)
}
fmt.Println("----------")

if resp.IsSuccess() { // status `code >= 200 and <= 299` is considered as success
	// Must have been marshaled to user if no error returned before
    fmt.Printf("%s's blog is %s\n", user.Name, user.Blog)
} else if resp.IsError() { // status `code >= 400` is considered as error
	// Must have been marshaled to errMsg if no error returned before
    fmt.Println("got error:", errMsg.Message) 
} else {
    log.Fatal("unknown http status:", resp.Status)
}
/* Output
{"login":"imroc","id":7448852,"node_id":"MDQ6VXNlcjc0NDg4NTI=","avatar_url":"https://avatars.githubusercontent.com/u/7448852?v=4","gravatar_id":"","url":"https://api.github.com/users/imroc","html_url":"https://github.com/imroc","followers_url":"https://api.github.com/users/imroc/followers","following_url":"https://api.github.com/users/imroc/following{/other_user}","gists_url":"https://api.github.com/users/imroc/gists{/gist_id}","starred_url":"https://api.github.com/users/imroc/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/imroc/subscriptions","organizations_url":"https://api.github.com/users/imroc/orgs","repos_url":"https://api.github.com/users/imroc/repos","events_url":"https://api.github.com/users/imroc/events{/privacy}","received_events_url":"https://api.github.com/users/imroc/received_events","type":"User","site_admin":false,"name":"roc","company":"Tencent","blog":"https://imroc.cc","location":"China","email":null,"hireable":true,"bio":"I'm roc","twitter_username":"imrocchan","public_repos":129,"public_gists":0,"followers":362,"following":151,"created_at":"2014-04-30T10:50:46Z","updated_at":"2022-01-24T23:32:53Z"}
----------
roc's blog is https://imroc.cc
*/

// Or you can also unmarshal response later
if resp.IsSuccess() {
    err = resp.Unmarshal(user)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s's blog is %s\n", user.Name, user.Blog)
} else {
    fmt.Println("bad response:", resp)
}

// Also, you can get the raw response and Unmarshal by yourself
yaml.Unmarshal(resp.Bytes())
```

**Customize JSON&XML Marshal/Unmarshal**

```go
// Example of registering json-iterator
import jsoniter "github.com/json-iterator/go"

json := jsoniter.ConfigCompatibleWithStandardLibrary

client := req.C().
	SetJsonMarshal(json.Marshal).
	SetJsonUnmarshal(json.Unmarshal)

// Similarly, XML functions can also be customized
client.SetXmlMarshal(xmlMarshalFunc).SetXmlUnmarshal(xmlUnmarshalFunc)
```

**Disable Auto-Read Response Body**

Response body will be read into memory if it's not a download request by default, you can disable it if you want (normally you don't need to do this).

```go
client.DisableAutoReadResponse()

resp, err := client.R().Get(url)
if err != nil {
	log.Fatal(err)
}
io.Copy(dst, resp.Body)
```

## <a name="Cert">Custom Certificates</a>

```go
client := req.R()

// Set root cert and client cert from file path
client.SetRootCertsFromFile("/path/to/root/certs/pemFile1.pem", "/path/to/root/certs/pemFile2.pem", "/path/to/root/certs/pemFile3.pem"). // Set root cert from one or more pem files
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
client.SetCerts(cert1, cert2, cert3) 
```

## <a name="Auth">Basic Auth and Bearer Token</a>

```go
client := req.C()

// Set basic auth for all request
client.SetCommonBasicAuth("imroc", "123456")

// Set bearer token for all request
client.SetCommonBearerAuthToken("MDc0ZTg5YmU4Yzc5MjAzZGJjM2ZiMzkz")

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

// You can also use io.Reader to upload
avatarImgFile, _ := os.Open("avatar.png")
client.R().SetFileReader("avatar", "avatar.png", avatarImgFile).Post(url)
*/
```

## <a name="AutoDecode">Auto-Decode</a>

`Req` detect the charset of response body and decode it to utf-8 automatically to avoid garbled characters by default.

Its principle is to detect whether `Content-Type` header at first, if it's not the text content type (json, xml, html and so on), `req` will not try to decode. If it is, then `req` will try to find the charset information, if it's not included in the header, it will try to sniff the body's content to determine the charset, if found and is not utf-8, then decode it to utf-8 automatically, if the charset is not sure, it will not decode, and leave the body untouched.

You can also disable if you don't need or care a lot about performance:

```go
client.DisableAutoDecode()
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

## <a name="Middleware">Request and Response Middleware</a>

```go
client := req.C()

// Registering Request Middleware
client.OnBeforeRequest(func(c *req.Client, r *req.Request) error {
	// You can access Client and current Request object to do something
	// as you need

    return nil  // return nil if it is success
  })

// Registering Response Middleware
client.OnAfterResponse(func(c *req.Client, r *req.Response) error {
    // You can access Client and current Response object to do something
    // as you need

    return nil  // return nil if it is success
  })
```

## <a name="Redirect">Redirect Policy</a>

```go
client := req.C().EnableDumpOnlyRequest()

client.SetRedirectPolicy(
    // Only allow up to 5 redirects
    req.MaxRedirectPolicy(5),
    // Only allow redirect to same domain.
    // e.g. redirect "www.imroc.cc" to "imroc.cc" is allowed, but "google.com" is not
    req.SameDomainRedirectPolicy(),
)

client.SetRedirectPolicy(
    // Only *.google.com/google.com and *.imroc.cc/imroc.cc is allowed to redirect
    req.AllowedDomainRedirectPolicy("google.com", "imroc.cc"),
    // Only allow redirect to same host.
    // e.g. redirect "www.imroc.cc" to "imroc.cc" is not allowed, only "www.imroc.cc" is allowed
    req.SameHostRedirectPolicy(),
)

// All redirect is not allowd
client.SetRedirectPolicy(req.NoRedirectPolicy())

// Or customize the redirect with your own implementation
client.SetRedirectPolicy(func(req *http.Request, via []*http.Request) error {
    // ...
})
```

## <a name="Proxy">Proxy</a>

`Req` use proxy `http.ProxyFromEnvironment` by default, which will read the `HTTP_PROXY/HTTPS_PROXY/http_proxy/https_proxy` environment variable, and setup proxy if environment variable is been set. You can customize it if you need:

```go
// Set proxy from proxy url
client.SetProxyURL("http://myproxy:8080")

// Custmize the proxy function with your own implementation
client.SetProxy(func(request *http.Request) (*url.URL, error) {
    //...
})

// Disable proxy
client.SetProxy(nil)
```

## <a name="TODO">TODO List</a>

* [ ] Add tests.
* [ ] Wrap more transport settings into client.
* [ ] Support retry.
* [ ] Support unix socket.
* [ ] Support h2c.

## <a name="License">License</a>

`Req` released under MIT license, refer [LICENSE](LICENSE) file.
