<p align="center">
    <h1 align="center">Req</h1>
    <p align="center">Simple Go HTTP client with Black Magic (Less code and More efficiency).</p>
    <p align="center">
        <a href="https://github.com/imroc/req/actions/workflows/ci.yml?query=branch%3Amaster"><img src="https://github.com/imroc/req/actions/workflows/ci.yml/badge.svg" alt="Build Status"></a>
        <a href="https://codecov.io/gh/imroc/req/branch/master"><img src="https://codecov.io/gh/imroc/req/branch/master/graph/badge.svg" alt="Code Coverage"></a>
        <a href="https://goreportcard.com/report/github.com/imroc/req/v3"><img src="https://goreportcard.com/badge/github.com/imroc/req/v3" alt="Go Report Card"></a>
        <a href="https://pkg.go.dev/github.com/imroc/req/v3"><img src="https://pkg.go.dev/badge/github.com/imroc/req/v3.svg"></a>
        <a href="LICENSE"><img src="https://img.shields.io/github/license/imroc/req.svg" alt="License"></a>
        <a href="https://github.com/imroc/req/releases"><img src="https://img.shields.io/github/v/release/imroc/req?display_name=tag&sort=semver" alt="GitHub Releases"></a>
        <a href="https://github.com/avelino/awesome-go"><img src="https://awesome.re/mentioned-badge.svg" alt="Mentioned in Awesome Go"></a>
    </p> 
</p>

## News

Brand-New version v3 is released, which is completely rewritten, bringing revolutionary innovations and many superpowers, try and enjoy :)

If you want to use the older version, check it out on [v1 branch](https://github.com/imroc/req/tree/v1).

> v2 is a transitional version, due to some breaking changes were introduced during optmize user experience

## Documentation

Full documentation is available on the [Req Official Website](https://req.cool/).

## <a name="Features">Features</a>

* Simple and chainable methods for both client-level and request-level settings, and the request-level setting takes precedence if both are set.
* Powerful and convenient debug utilites, including debug logs, performance traces, and even dump the complete request and response content (see [Debugging - Dump/Log/Trace](#Debugging)).
* Easy making HTTP test with code instead of tools like curl or postman, `req` provide global wrapper methods and `MustXXX` to test API with minimal code (see [Quick HTTP Test](#Test)).
* Works fine with both `HTTP/2` and `HTTP/1.1`, which `HTTP/2` is preferred by default if server support, and you can also force `HTTP/1.1` if you want (see [HTTP2 and HTTP1](#HTTP2-HTTP1)).
* Detect the charset of response body and decode it to utf-8 automatically to avoid garbled characters by default (see [Auto-Decode](#AutoDecode)).
* Automatic marshal and unmarshal for JSON and XML content type and fully customizable (see [Body and Marshal/Unmarshal](#Body)).
* Exportable `Transport`, easy to integrate with existing `http.Client`, debug APIs with minimal code change.
* Easy [Download and Upload](#Download-Upload).
* Easy set header, cookie, path parameter, query parameter, form data, basic auth, bearer token for both client and request level.
* Easy set timeout, proxy, certs, redirect policy, cookie jar, compression, keepalives etc for client.
* Support middleware before request sent and after got response (see [Request and Response Middleware](#Middleware)).

## <a name="Get-Started">Get Started</a>

**Install**

``` sh
go get github.com/imroc/req/v3
```

**Import**

```go
import "github.com/imroc/req/v3"
```

**Basic Usage**

```go
// For test, you can create and send a request with the global default
// client, use DevMode to see all details, try and suprise :)
req.DevMode()
req.Get("https://httpbin.org/get")

// In production, create a client explicitly and reuse it to send all requests
// Create and send a request with the custom client and settings.
client := req.C(). // Use C() to create a client and set with chainable client settings.
    SetUserAgent("my-custom-client").
    SetTimeout(5 * time.Second).
    DevMode()
resp, err := client.R(). // Use R() to create a request and set with chainable request settings.
    SetHeader("Accept", "application/vnd.github.v3+json").
    SetPathParam("username", "imroc").
    SetQueryParam("page", "1").
    SetResult(&result). // Unmarshal response into struct automatically.
    Get("https://api.github.com/users/{username}/repos")
```

**Videos**

* [Get Started With Req](https://www.youtube.com/watch?v=k47i0CKBVrA) (English, Youtube)
* [快速上手 req](https://www.bilibili.com/video/BV1Xq4y1b7UR) (Chinese, BiliBili)

**More**

Check more introduction, tutorials, examples and API references on the [official website](https://req.cool/).

## <a name="License">License</a>

`Req` released under MIT license, refer [LICENSE](LICENSE) file.
