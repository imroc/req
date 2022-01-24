# req

[![GoDoc](https://pkg.go.dev/badge/github.com/imroc/req.svg)](https://pkg.go.dev/github.com/imroc/req)

A golang http request library for humans.

* [Features](#Features)
* [Quick Start](#Quick-Start)
* [API Design](#API-Design)
* [Examples](#Examples)
  
## <a name="Features">Features</a>

* Simple and chainable methods for client and request settings, rich syntax sugar, less code and more efficiency.
* Automatically detect charset and decode it to utf-8.
* Powerful debugging capabilities (logging, tracing, and even dump the requests and responses' content).
* All settings can be changed dynamically, making it possible to debug in the production environment.
* Easy to integrate with existing code, just replace the Transport of existing http.Client, then you can dump content as req to debug APIs.

## <a name="Quick-Start">Quick Start</a>

**Install**

``` sh
go get github.com/imroc/req/v2@v2.0.0-alpha.7
```

**Import**

```go
import "github.com/imroc/req/v2"
```

**Simple GET**

```go
// Create and send a request with the global default client
req.DevMode()
resp, err := req.Get("https://api.github.com/users/imroc")

// Create and send a request with the custom client
client := req.C().DevMode()
resp, err := client.R().Get("https://api.github.com/users/imroc")
```

**Client Settings**

```go
// Create a client with custom client settings using chainable method
client := req.C().SetUserAgent("my-custom-client").SetTimeout(5 * time.Second).DevMode()

// You can also configure the global client using the same chainable method,
// req wraps global method for default client 
req.SetUserAgent("my-custom-client").SetTimeout(5 * time.Second).DevMode()
```

**Request Settings**

```go
// Use client to create a request with custom request settings
client := req.C().DevMode()
var result Result
resp, err := client.R().
    SetHeader("Accept", "application/json").
    SetQueryParam("page", "1").
    SetPathParam("userId", "imroc").
    SetResult(&result).
    Get(url)

// You can also create a request using global client using the same chaining method
resp, err := req.R().
    SetHeader("Accept", "application/json").
    SetQueryParam("page", "1").
    SetPathParam("userId", "imroc").
    SetResult(&result).
    Get(url)

// You can even also create a request without calling R(), cuz req
// wraps global method for request, and create a request using
// the default client automatically.
resp, err := req.SetHeader("Accept", "application/json").
    SetQueryParam("page", "1").
    SetPathParam("userId", "imroc").
    SetResult(&result).
    Get(url)
```

## <a name="API-Design">API Design</a>

**Settings and Chainable Methods**

`Request` and `Client` is the most important object, and both need to support a lot of settings, req provide rich chainable methods out of the box, making it very simple and intuitive to initiate any request you want.

**Global Wrapper for Testing Purposes**

`req` wraps global methods of `Client` and `Request` for testing purposes, so that you don't even need to create the client or request explicitly, just test API with minimal code like this:

```go
req.DevMode().SetCommonBasicAuth("imroc", "123456")
req.SetBodyJsonString(`{"nickname":"roc", "email":"roc@imroc.cc"}`).Post("https://api.exmaple.com/profile")
```

**Conmmon Methods and Override**

There are some similar methods between `Client` and `Request`, the pattern is like `Request.SetXXX` corresponding to `Client.SetCommonXXX`, client settings take effect for all requests, but will be overridden if the request sets the same setting.

The common methods list is:

* `Request.SetHeader` vs `Client.SetCommonHeader`
* `Request.SetHeaders` vs `Client.SetCommonHeaders`
* `Request.SetCookie` vs `Client.SetCommonCookie`
* `Request.SetCookies` vs `Client.SetCommonCookies`
* `Request.SetBasicAuth` vs `Client.SetCommonBasicAuth`
* `Request.SetBearerToken` vs `Client.SetCommonBearerToken`
* `Request.SetQueryParam` vs `Client.SetCommonQueryParam`
* `Request.SetQueryParams` vs `Client.SetCommonQueryParams`
* `Request.SetQueryString` vs `Client.SetCommonQueryString`

## <a name="Examples">Examples</a>

* [Debug](#Debug)
* [Set Path Parameter](#PathParam)
* [Set Query Parameter](#QueryParam)
* [Set Header](#Header)
* [Set Cookie](#Cookie)
* [Set Cert](#Cert)
* [Set Basic Auth and Bearer Token](#Auth)

You can find more complete and runnable examples [here](examples).

### <a name="Debug">Debug</a>

**Dump the content of request and response**

```go
// Set EnableDump to true, dump all content to stdout by default,
// including both the header and body of all request and response
client := req.C().EnableDump(true)
client.R().Get("https://api.github.com/users/imroc")

/* Output
:authority: api.github.com
:method: GET
:path: /users/imroc
:scheme: https
user-agent: req/v2 (https://github.com/imroc/req)
accept-encoding: gzip

:status: 200
server: GitHub.com
date: Fri, 21 Jan 2022 09:31:43 GMT
content-type: application/json; charset=utf-8
cache-control: public, max-age=60, s-maxage=60
vary: Accept, Accept-Encoding, Accept, X-Requested-With
etag: W/"fe5acddc5c01a01153ebc4068a1f067dadfa7a7dc9a025f44b37b0a0a50e2c55"
last-modified: Thu, 08 Jul 2021 12:11:23 GMT
x-github-media-type: github.v3; format=json
access-control-expose-headers: ETag, Link, Location, Retry-After, X-GitHub-OTP, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Used, X-RateLimit-Resource, X-RateLimit-Reset, X-OAuth-Scopes, X-Accepted-OAuth-Scopes, X-Poll-Interval, X-GitHub-Media-Type, X-GitHub-SSO, X-GitHub-Request-Id, Deprecation, Sunset
access-control-allow-origin: *
strict-transport-security: max-age=31536000; includeSubdomains; preload
x-frame-options: deny
x-content-type-options: nosniff
x-xss-protection: 0
referrer-policy: origin-when-cross-origin, strict-origin-when-cross-origin
content-security-policy: default-src 'none'
content-encoding: gzip
x-ratelimit-limit: 60
x-ratelimit-remaining: 59
x-ratelimit-reset: 1642761103
x-ratelimit-resource: core
x-ratelimit-used: 1
accept-ranges: bytes
content-length: 486
x-github-request-id: AF10:6205:BA107D:D614F2:61EA7D7E

{"login":"imroc","id":7448852,"node_id":"MDQ6VXNlcjc0NDg4NTI=","avatar_url":"https://avatars.githubusercontent.com/u/7448852?v=4","gravatar_id":"","url":"https://api.github.com/users/imroc","html_url":"https://github.com/imroc","followers_url":"https://api.github.com/users/imroc/followers","following_url":"https://api.github.com/users/imroc/following{/other_user}","gists_url":"https://api.github.com/users/imroc/gists{/gist_id}","starred_url":"https://api.github.com/users/imroc/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/imroc/subscriptions","organizations_url":"https://api.github.com/users/imroc/orgs","repos_url":"https://api.github.com/users/imroc/repos","events_url":"https://api.github.com/users/imroc/events{/privacy}","received_events_url":"https://api.github.com/users/imroc/received_events","type":"User","site_admin":false,"name":"roc","company":"Tencent","blog":"https://imroc.cc","location":"China","email":null,"hireable":true,"bio":"I'm roc","twitter_username":"imrocchan","public_repos":128,"public_gists":0,"followers":362,"following":151,"created_at":"2014-04-30T10:50:46Z","updated_at":"2021-07-08T12:11:23Z"}
*/
	
// Dump header content asynchronously and save it to file
client.EnableDumpOnlyHeader(). // Only dump the header of request and response
    EnableDumpAsync(). // Dump asynchronously to improve performance
    EnableDumpToFile("reqdump.log") // Dump to file without printing it out
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

**Debug Logging**

```go
// Logging is enabled by default, but only output warning and error message to stdout.
// EnableDebug set to true to enable debug level message logging.
client := req.C().EnableDebug(true)
client.R().Get("https://api.github.com/users/imroc")
// Output
// 2022/01/23 14:33:04.755019 DEBUG [req] GET https://api.github.com/users/imroc

// SetLogger with nil to disable all log
client.SetLogger(nil)

// Or customize the logger with your own implementation.
client.SetLogger(logger)
```

**DevMode**

If you want to enable all debug features (dump, debug logging and tracing), just call `DevMode()`:

```go
client := req.C().DevMode()
client.R().Get("https://imroc.cc")
```

### <a name="PathParam">Set Path Parameter</a>

Use `PathParam` to replace variable in the url path:

```go
client := req.C().DevMode()

client.R().
    SetPathParam("owner", "imroc"). // Set a path param, which will replace the vairable in url path
    SetPathParams(map[string]string{ // Set multiple path params at once 
        "repo": "req",
        "path": "README.md",
    }).Get("https://api.github.com/repos/{owner}/{repo}/contents/{path}")
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

### <a name="QueryParam">Set Query Parameter</a>

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

### <a name="Header">Set Header</a>

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

### <a name="Cookie">Set Cookie</a>

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

### <a name="Cert">Set Cert</a>

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

### <a name="Auth">Set Basic Auth and Bearer Token</a>

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

## License

Req released under MIT license, refer [LICENSE](LICENSE) file.