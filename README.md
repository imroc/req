# req
[![GoDoc](https://godoc.org/github.com/imroc/req?status.svg)](https://godoc.org/github.com/imroc/req)

A golang http request library.



Features
========

- light weight
- simple
- easy to set headers and params
- easy to deserialize response ([`ToJSON`, `ToXML`](#ToJSON-ToXML))
- easy to debug and logging ([print info](#Debug-Logging))
- easy to upload and download
- easy to manage cookie
- easy to set up proxy
- easy to set timeout

Install
=======
``` sh
go get github.com/imroc/req
```

Quick Start
=======
`Request = Method + Url [+ Options]`  
the "`Options`" can be headers, params, files or body etc
``` go
// execute a request
func Do(method, url string, v ...interface{}) (*Req, error)
// wraps Do
func Get(url string, v ...interface{}) (*Req, error)
func Post(url string, v ...interface{}) (*Req, error)
......
```

Examples
=======
[Basic](#Basic)  
[Set Header](#Set-Header)  
[Set Param](#Set-Param)  
[Set Body](#Set-Body)  
[Debug / Logging](#Debug-Logging)  
[ToJSON / ToXML](#ToJSON-ToXML)  
[Upload](#Upload)  
[Download](#Download)  
[Cookie](#Cookie)  
[Set Timeout](#Set-Timeout)  
[Set Proxy](#Set-Proxy)  
[Set Client](#Set-Client)  

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

## <a name="Debug-Logging">Debug / Logging</a>
use `%+v` format to print info in detail
``` go
r, _ := req.Post(url, header, param)
log.Printf("%+v", r)
/*
	POST http://foo.bar/api HTTP/1.1
	Authorization:Basic YWRtaW46YWRtaW4=
	Accept:application/json
	Content-Type:application/x-www-form-urlencoded

	city=Chengdu&cmd=list_gopher

	HTTP/1.1 200 OK
	Content-Type:application/json; charset=UTF-8
	Date:Wed, 03 May 2017 09:39:27 GMT
	Content-Length:39

	{"code":0,"name":["imroc","yulibaozi"]}
*/
```
use `%v` format to print info simple
``` go
r, _ := req.Get(url, param)
log.Printf("%v", r) // GET http://foo.bar/api?name=roc&cmd=add {"code":"0","msg":"success"}
```
and the `%-v` format is similar to `%v`, the only difference is that it removes all blank characters, keep content minimal and in one line, it is useful while logging.

## <a name="ToJSON-ToXML">ToJSON / ToXML</a>
``` go
r, _ := req.Get(url)
r.ToJSON(&foo)
r, _ = req.Post(url, req.BodyXML(&bar))
r.ToXML(&baz)
```

## <a name="Upload">Upload</a>
specify filename
``` go
req.Post(url, req.File("imroc.png"), req.File("/bin/sh"))
```
match pattern
``` go
req.Post(url, req.FileGlob("/usr/*/bin/go*"))
```
use `req.FileUpload`
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

## <a name="Set-Client">Set Client</a>
Use `req.SetClient` to change the default underlying `*http.Client`
``` go
req.SetClient(client)
```