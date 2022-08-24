# req

<p align="center">
    <p align="center"><img src="https://req.cool/images/req.png"></p>
    <p align="center"><strong>Simple Go HTTP client with Black Magic</strong></p>
    <p align="center">
        <a href="https://github.com/imroc/req/actions/workflows/ci.yml?query=branch%3Amaster"><img src="https://github.com/imroc/req/actions/workflows/ci.yml/badge.svg" alt="Build Status"></a>
        <a href="https://goreportcard.com/report/github.com/imroc/req/v3"><img src="https://goreportcard.com/badge/github.com/imroc/req/v3" alt="Go Report Card"></a>
        <a href="https://pkg.go.dev/github.com/imroc/req/v3"><img src="https://pkg.go.dev/badge/github.com/imroc/req/v3.svg"></a>
        <a href="LICENSE"><img src="https://img.shields.io/github/license/imroc/req.svg" alt="License"></a>
        <a href="https://github.com/imroc/req/releases"><img src="https://img.shields.io/github/v/release/imroc/req?display_name=tag&sort=semver" alt="GitHub Releases"></a>
        <a href="https://github.com/avelino/awesome-go"><img src="https://awesome.re/mentioned-badge.svg" alt="Mentioned in Awesome Go"></a>
    </p> 
</p>

## Documentation

Full documentation is available on the official website: https://req.cool.

## <a name="Features">Features</a>

* **Simple and Powerful**: Simple and easy to use, providing rich client-level and request-level settings, all of which are intuitive and chainable methods.
* **Easy Debugging**: Powerful and convenient debug utilities, including debug logs, performance traces, and even dump the complete request and response content (see [Debugging](https://req.cool/docs/tutorial/debugging/)).
* **Easy API Testing**: API testing can be done with minimal code, no need to explicitly create any Request or Client, or even to handle errors (See [Quick HTTP Test](https://req.cool/docs/tutorial/quick-test/))
* **Smart by Default**: Detect and decode to utf-8 automatically if possible to avoid garbled characters (See [Auto Decode](https://req.cool/docs/tutorial/auto-decode/)), marshal request body and unmarshal response body automatically according to the Content-Type.
* **Support Multiple HTTP Versions**: Support `HTTP/1.1`, `HTTP/2`, and `HTTP/3`, and can automatically detect the server side and select the optimal HTTP version for requests, you can also force the protocol if you want (See [Force HTTP version](https://req.cool/docs/tutorial/force-http-version/)).
* **Support Retry**: Support automatic request retry and is fully customizable (See [Retry](https://req.cool/docs/tutorial/retry/)).
* **Easy Download and Upload**: You can download and upload files with simple request settings, and even set a callback to show real-time progress (See [Download](https://req.cool/docs/tutorial/download/) and [Upload](https://req.cool/docs/tutorial/upload/)).
* **Exportable**: `req.Transport` is exportable. Compared with `http.Transport`, it also supports HTTP3, dump content, middleware, etc. It can directly replace the Transport of `http.Client` in existing projects, and obtain more powerful functions with minimal code change.
* **Extensible**: Support Middleware for Request, Response, Client and Transport (See [Request and Response Middleware](https://req.cool/docs/tutorial/middleware-for-request-and-response/)) and [Client and Transport Middleware](https://req.cool/docs/tutorial/middleware-for-client-and-transport/)).

## <a name="Get-Started">Get Started</a>

**Install**

You first need [Go](https://go.dev/) installed (version 1.16+ is required), then you can use the below Go command to install req:

``` sh
go get github.com/imroc/req/v3
```

**Import**

Import req to your code:

```go
import "github.com/imroc/req/v3"
```

**Basic Usage**

```go
// For testing, you can create and send a request with the global wrapper methods
// that use the default client behind the scenes to initiate the request (you can
// just treat package name `req` as a Client or Request, no need to create any client
// or Request explicitly).
req.DevMode() //  Use Client.DevMode to see all details, try and surprise :)
req.Get("https://httpbin.org/get") // Use Request.Get to send a GET request.

// In production, create a client explicitly and reuse it to send all requests
client := req.C(). // Use C() to create a client and set with chainable client settings.
    SetUserAgent("my-custom-client").
    SetTimeout(5 * time.Second).
    DevMode()
resp, err := client.R(). // Use R() to create a request and set with chainable request settings.
    SetHeader("Accept", "application/vnd.github.v3+json").
    SetPathParam("username", "imroc").
    SetQueryParam("page", "1").
    SetResult(&result). // Unmarshal response into struct automatically if status code >= 200 and <= 299.
    SetError(&errMsg). // Unmarshal response into struct automatically if status code >= 400.
    EnableDump(). // Enable dump at request level to help troubleshoot, log content only when an unexpected exception occurs.
    Get("https://api.github.com/users/{username}/repos")
if err != nil {
    // Handle error.
    // ...
    return
}
if resp.IsSuccess() {
    // Handle result.
    // ...
    return
}
if resp.IsError() {
    // Handle errMsg.	
    // ...
    return
}
// Handle unexpected response (corner case).
err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())
// ...
```

You can also use another style if you want:

```go
resp := client.Get("https://api.github.com/users/{username}/repos"). // Create a GET request with specified URL.
    SetHeader("Accept", "application/vnd.github.v3+json").
    SetPathParam("username", "imroc").
    SetQueryParam("page", "1").
    SetResult(&result).
    SetError(&errMsg).
    EnableDump().
    Do() // Send request with Do.

if resp.Err != nil {
    // ...
}
// ...
```

**Videos**

The following is a series of video tutorials for req:

* [Youtube Play List](https://www.youtube.com/watch?v=Dy8iph8JWw0&list=PLnW6i9cc0XqlhUgOJJp5Yf1FHXlANYMhF&index=2)
* [BiliBili 播放列表](https://www.bilibili.com/video/BV14t4y1J7cm) (Chinese)

**More**

Check more introduction, tutorials, examples, best practices and API references on the [official website](https://req.cool/).

## Contributing

If you have a bug report or feature request, you can [open an issue](https://github.com/imroc/req/issues/new), and [pull requests](https://github.com/imroc/req/pulls) are also welcome.

## Contact

If you have questions, feel free to reach out to us in the following ways:

* [Github Discussion](https://github.com/imroc/req/discussions)
* [Slack](https://imroc-req.slack.com/archives/C03UFPGSNC8) | [Join](https://slack.req.cool/)
* QQ Group (Chinese): 621411351 - <a href="https://qm.qq.com/cgi-bin/qm/qr?k=P8vOMuNytG-hhtPlgijwW6orJV765OAO&jump_from=webapi"><img src="https://pub.idqqimg.com/wpa/images/group.png"></a>

## <a name="License">License</a>

`Req` released under MIT license, refer [LICENSE](LICENSE) file.
