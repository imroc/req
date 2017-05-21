# req
[![GoDoc](https://godoc.org/github.com/imroc/req?status.svg)](https://godoc.org/github.com/imroc/req)

A golang http request library for human



Features
========

- Light weight
- Simple
- Easy play with JSON and XML
- Easy for debug and logging
- Easy file uploads and downloads
- Easy manage cookie
- Easy set up proxy
- Easy set timeout
- Easy customize http client

Install
=======
``` sh
go get github.com/imroc/req
```

Quick Start
=======
Only url is required, others are optional, like headers, params, files or body etc
``` go
func Get(url string, v ...interface{}) (*Req, error)
func Post(url string, v ...interface{}) (*Req, error)
......
```
The struct `Req` hold both request and response infomation.

Examples
=======
[Basic](#Basic)  
[Set Header](#Set-Header)  
[Set Param](#Set-Param)  
[Set Body](#Set-Body)  
[Debug & Logging](#Debug-Logging)  
[ToJSON & ToXML](#ToJSON-ToXML)  
[Get *http.Response](#Response)  
[Upload](#Upload)  
[Download](#Download)  
[Cookie](#Cookie)  
[Set Timeout](#Set-Timeout)  
[Set Proxy](#Set-Proxy)  
[Customize Client](#Customize-Client)  

## <a name="Basic">Basic</a>
``` go
header := req.Header{
	"Accept":        "application/json",
	"Authorization": "Basic YWRtaW46YWRtaW4=",
}
param := req.Param{
	"name": "imroc",
	"cmd":  "add",
}
// only url is required, others are optional.
r, err = req.Post("http://foo.bar/api", header, param)
if err != nil {
	log.Fatal(err)
}
r.ToJSON(&foo)       // response => struct/map
log.Printf("%+v", r) // print info (try it, you may surprise) 
```

## <a name="Set-Header">Set Header</a>
use `req.Header`
``` go
authHeader := req.Header{
	"Accept":        "application/json",
	"Authorization": "Basic YWRtaW46YWRtaW4=",
}
req.Get("https://www.baidu.com", authHeader, req.Header{"User-Agent": "V1.1"})
```
use `http.Header`
``` go
header := make(http.Header)
header.Set("Accept", "application/json")
r, err := req.Get("https://www.baidu.com", header)
```

## <a name="Set-Param">Set Param</a>
use `req.Param`
``` go
param := req.Param{
	"id":  "imroc",
	"pwd": "roc",
}
req.Get("http://foo.bar/api", param) // http://foo.bar/api?id=imroc&pwd=roc
req.Post(url, param)                  // body => id=imroc&pwd=roc
```
use `req.QueryParam` force to append params to the url
``` go
req.Post("http://foo.bar/api", req.Param{"name": "roc", "age": "22"}, req.QueryParam{"access_token": "fedledGF9Hg9ehTU"})
/*
POST /api?access_token=fedledGF9Hg9ehTU HTTP/1.1
Host: foo.bar
User-Agent: Go-http-client/1.1
Content-Length: 15
Content-Type: application/x-www-form-urlencoded;charset=UTF-8
Accept-Encoding: gzip

age=22&name=roc
*/
```

## <a name="Set-Body">Set Body</a>
put `string`, `[]byte` and `io.Reader` as body directly.
``` go
req.Post(url, "id=roc&cmd=query")
```
put xml and json body
``` go
req.Post(url, req.BodyJSON(&foo))
req.Post(url, req.BodyXML(&bar))
```

## <a name="Debug-Logging">Debug & Logging</a>
Enable debug mode
``` go
req.Debug = true
req.Get(url) // it will print the debug info, try it :)
```
Monitor speed
``` go
req.ShowCost = true // show how many time costed by every request if you print it or debug mode is enabled
r,_ := req.Get(url)
log.Println(r) // http://foo.bar/api 3.260802ms {"code":0 "msg":"success"}
if r.Cost() > 3 * time.Second { // check cost
	log.Println("WARN: slow request:",r)
}
```
Use `%+v` format to print debug info
``` go
r, _ := req.Post(url, header, param)
log.Printf("%+v", r)
```
Use `%v` format to print info simple
``` go
r, _ := req.Get(url, param)
log.Printf("%v", r) // GET http://foo.bar/api?name=roc&cmd=add {"code":"0","msg":"success"}
```
and the `%-v` format is similar to `%v`, the only difference is that it always keep the content in one line, it is useful while logging.

## <a name="ToJSON-ToXML">ToJSON & ToXML</a>
``` go
r, _ := req.Get(url)
r.ToJSON(&foo)
r, _ = req.Post(url, req.BodyXML(&bar))
r.ToXML(&baz)
```

## <a name="Response">Get *http.Response</a>
```go
// func (r *Req) Response() *http.Response
r, _ := req.Get(url)
resp := r.Response()
fmt.Println(resp.StatusCode)
```

## <a name="Upload">Upload</a>
Use `req.File` to match files
``` go
req.Post(url, req.File("imroc.png"), req.File("/Users/roc/Pictures/*.png"))
```
Use `req.FileUpload` to fully control
``` go
file, _ := os.Open("imroc.png")
req.Post(url, req.FileUpload{
	File:      file,
	FieldName: "file",
	FileName:  "avatar.png",
})
```

## <a name="Download">Download</a>
``` go
r, _ := req.Get(url)
r.ToFile("imroc.png")
```

## <a name="Cookie">Cookie</a>
By default, the underlying `*http.Client` will manage your cookie(send cookie header to server automatically if server has set a cookie for you), you can disable it by calling this function :
``` go
req.EnableCookie(false)
```
and you can set cookie in request just using `*http.Cookie`
``` go
cookie := new(http.Cookie)
......
req.Get(url, cookie)
```

## <a name="Set-Timeout">Set Timeout</a>
``` go
req.SetTimeout(50 * time.Second)
```

## <a name="Set-Proxy">Set Proxy</a>
By default, req use proxy from system environment if `http_proxy` or `https_proxy` is specified, you can set a custom proxy or disable it by set `nil`
``` go
req.SetProxy(func(r *http.Request) (*url.URL, error) {
	if strings.Contains(r.URL.Hostname(), "google") {
		return url.Parse("http://my.vpn.com:23456")
	}
	return nil, nil
})
```
Set a simple proxy (use fixed proxy url for every request)
``` go
req.SetProxyUrl("http://my.proxy.com:23456")
```

## <a name="Customize-Client">Customize Client</a>
Use `req.SetClient` to change the default underlying `*http.Client`
``` go
req.SetClient(client)
```
Only specify client for some request
``` go
client := &http.Client{Timeout: 30 * time.Second}
req.Get(url, client)
```
Change some properties of default client you want
``` go
req.Client().Jar, _ = cookiejar.New(nil)
trans, _ := req.Client().Transport.(*http.Transport)
trans.MaxIdleConns = 20
trans.TLSHandshakeTimeout = 20 * time.Second
trans.DisableKeepAlives = true
trans.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
```