req
==============
req is a super light weight and super easy-to-use  http request library.


# Quick Start
## Install
``` sh
go get github.com/imroc/req
```

## Basic 
``` go
req.Get(url).String() // get response as string
req.Post(url).Body(body).ToJson(&foo) // set request body as string or []byte, get response unmarshal to struct.
fmt.Println(req.Get("http://api.foo")) // GET http://api.foo {"code":0,"msg":"success"}
/*
	POST http://api.foo HTTP/1.1

	Content-Type:application/x-www-form-urlencoded
	User-Agent:Chrome/57.0.2987.110

	id=1

	HTTP/1.1 200 OK
	Server:nginx
	Set-Cookie:bla=3154899087195606076; expires=Wed, 29-Mar-17 09:18:18 GMT; domain=api.foo; path=/
	Connection:keep-alive
	Content-Type:application/json

	{"code":0,"data":{"name":"req"}}
*/
fmt.Printf("%+v",req.Post("http://api.foo").Param("id","1").Header("User-Agent","Chrome/57.0.2987.110"))
```

## Set Params And Headers
``` go
r := req.Get("http://api.foo/get")
r.Params(req.M{
	"p1": "1",
	"p2": "2",
})
r.Headers(req.M{
	"Referer":    "http://api.foo",
	"User-Agent": "Chrome/57.0.2987.110",
})
r.GetUrl() // http://api.foo/get?p1=1&p2=2

r = req.Post("http://api.foo/post").Param("p3", "3").Header("Referer", "http://api.foo")
r.GetBody() // p3=3
```

## Get Response
```go
r := req.Get(url)
r.Response()   // *req.Response
r.String()     // string
r.Bytes()      // []byte
r.ToJson(&foo) // json->struct
r.ToXml(&bar)  // xml->struct

// ReceiveXXX will return error if error happens during
// the request been executed.
_, err = r.ReceiveResponse()
_, err = r.ReceiveString()
_, err = r.ReceiveBytes()
```
**NOTE:** By default, the underlying request will be executed only once when you call methods to get response like above.
You can retry the request by calling `Do` method, which will always execute the request, or you can call `Undo`, making the request could be executed again when calling methods to get the response next time.

## Dump Info
Sometimes you might want to dump the detail about the http request and response for debug or logging reasons. 
There are several format to output these detail infomation.


#### Default Format
Use `%v` or `%s` to get the info in default format.
``` go
r := req.Get("http://api.foo/get")
log.Printf("%v", r) // GET http://api.foo/get {"success":true,"data":"hello req"}
r = req.Post("http://api.foo/post").Body(`{"uid":"1"}`)
log.Println(r) // POST http://api.foo/post {"uid":"1"} {"success":true,"data":{"name":"req"}}
```
**NOTE** it will add newline if possible, keep it looks pretty. 


#### All Info Format
Use `%+v` or `%+s` to get the maximal detail infomation.
``` go
r := req.Post("http://api.foo/post")
r.Header("Referer": "http://api.foo")
r.Params(req.M{
	"p1": "1",
	"p2": "2",
})
/*
	POST http://api.foo/post HTTP/1.1

	Referer:http://api.foo
	Content-Type:application/x-www-form-urlencoded

	p1=1&p2=2

	HTTP/1.1 200 OK
	Server:nginx
	Set-Cookie:bla=3154899087195606076; expires=Wed, 29-Mar-17 09:18:18 GMT; domain=api.foo; path=/
	Expires:Thu, 30 Mar 2017 09:18:13 GMT
	Cache-Control:max-age=86400
	Date:Wed, 29 Mar 2017 09:18:13 GMT
	Connection:keep-alive
	Accept-Ranges:bytes
	Content-Type:application/json

	{"code":0,"data":{"name":"req"}}
*/
log.Printf("%+v", r)
```
As you can see, it will output the request Method,URL,Proto,[Request Header],[Request Body],[Response Header],[Response Body]


#### Oneline Format
Use `%-v` or `%-s` keeps info in one line (delete all blank characters if possible), this is useful while logging.
``` go
r := req.Get("http://api.foo/get")
// it output every thing in one line, even if '\n' exsist in reqeust body or response body.
log.Printf("%-v\n",r) // GET http://api.foo/get {"code":3019,"msg":"system busy"}
```


#### Request Only Format (No Response Info)
Use `%r`, `%+r` or `%-r` only output the request itself, no response.
``` go
r := req.Post("https://api.foo").Body(`name=req`)
fmt.Printf("%r", r) // POST https://api.foo name=req
```
**NOTE** in other format, it will execute the underlying request to get response if the request is not executed yet, you can disable that by using "Request Only Format".


#### Response Only
You need get the *req.Response, use `%v`,`%s`,`%+v`,`%+s`,`%-v`,`%-s` to output formatted response info.
``` go
resp := req.Get("http://api.foo").Response()
log.Println(resp)
log.Printf("%-s", resp)
log.Printf("%+s", resp)
```

## Setting
**NOTE** All settings methods is prefixed with Set
#### Set Timeout
``` go
req.Get("http://api.foo").
	SetReadTimeout(40 * time.Second). // read timeout
	SetWriteTimeout(30 * time.Second). // write timeout
	SetDialTimeout(20 * time.Second).  // dial timeout
	SetTimeout(60 * time.Second).     // total timeout
	String()
```

#### Set Proxy
``` go
req.Get("http://api.foo").
	SetProxy(func(r *http.Request) (*url.URL, error) {
		return url.Parse("http://localhost:40012")
	}).String()
```

#### Set Insecure TLS (Skip Verify Certificate Chain And Host Name)
``` go
req.Get("https://api.foo").SetInsecureTLS(true).String()
```

TODO
