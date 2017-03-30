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
req.Get("http://api.io/get").String() // {"success":true,"data":"hello roc"}
req.Post("http://api.io/post").Body(`{"uid":"1"}`).String() // {"success":true,"data":{"name":"roc"}}
```

## Set Params And Headers
``` go
r := req.Get("http://api.io/get").Params(req.M{
	"p1": "1",
	"p2": "2",
}).Headers(req.M{
	"Referer":    "http://api.io",
	"User-Agent": "Chrome/57.0.2987.110",
})
r.GetUrl() // http://api.io/get?p1=1&p2=2

r = req.Post("http://api.io/post").Param("p3", "3").Header("Referer", "http://api.io")
r.GetBody() // p3=3
```

## Get Response
```go
r := req.Get("http://api.io/get")
fmt.Println(r)            // GET http://api.io/get {"success":true,"msg":"hello req"}
resp := r.Response()      // *req.Response
str := r.String()         // string
bs := r.Bytes()           // []byte
err := r.ToJson(&jsonObj) // json->struct
err = r.ToXml(&xmlObj)    // xml->struct

// ReceiveXXX will return error if error happens during
// the request been executed.
resp, err = r.ReceiveResponse()
str, err = r.ReceiveString()
bs, err = r.ReceiveBytes()
```
**NOTE:** By default, the underlying request will be executed only once when you call methods to get response like above.
You can retry the request by calling `Do` method, which will always execute the request, or you can call `Undo`, making the request could be executed again when calling methods to get the response next time.

## Dump Info
Sometimes you might want to dump the detail about the http request and response for debug or logging reasons. 
There are several format to output these detail infomation.


#### Default Format
Use `%v` or `%s` to get the info in default format.
``` go
r := req.Get("http://api.io/get")
log.Printf("%v", r) // GET http://api.io/get {"success":true,"data":"hello req"}
r = req.Post("http://api.io/post").Body(`{"uid":"1"}`)
log.Println(r) // POST http://api.io/post {"uid":"1"} {"success":true,"data":{"name":"req"}}
```
**NOTE** it will add newline if possible, keep it looks pretty. 


#### All Info Format
Use `%+v` or `%+s` to get the maximal detail infomation.
``` go
r := req.Post("http://api.io/post").Header("Referer": "http://api.io").Params(req.M{
	"p1": "1",
	"p2": "2",
})
/*
	POST http://api.io/post HTTP/1.1

	Referer:http://api.io
	Content-Type:application/x-www-form-urlencoded

	p1=1&p2=2

	HTTP/1.1 200 OK
	Server:nginx
	Set-Cookie:bla=3154899087195606076; expires=Wed, 29-Mar-17 09:18:18 GMT; domain=api.io; path=/
	Expires:Thu, 30 Mar 2017 09:18:13 GMT
	Cache-Control:max-age=86400
	Date:Wed, 29 Mar 2017 09:18:13 GMT
	Connection:keep-alive
	Accept-Ranges:bytes
	Content-Type:application/json

	{"success":true,"data":{"name":"req"}}
*/
log.Printf("%+v", r)
```
As you can see, it will output the request Method,URL,Proto,[Request Header],[Request Body],[Response Header],[Response Body]


#### Oneline Format
Use `%-v` or `%-s` keeps info in one line (delete all blank characters if possible), this is useful while logging.
``` go
r := req.Get("http://api.io/get")
// it output every thing in one line, even if '\n' exsist in reqeust body or response body.
log.Printf("%-v\n",r) // GET http://api.io/get {"success":false,"msg":"system busy"}
```


#### Request Only Format (No Response Info)
Use `%r`, `%+r` or `%-r` only output the request itself, no response.
``` go
r := req.Get("https://api.io")
fmt.Printf("%r", r) // GET https://api.io HTTP/1.1
```
**NOTE** in other format, it will execute the underlying request to get response if the request is not executed yet, you can disable that by using "Request Only Format".


#### Response Only
You need get the *req.Response, use `%v`,`%s`,`%+v`,`%+s`,`%-v`,`%-s` to output formatted response info.
``` go
resp := req.Get("http://api.io").Response()
log.Println(resp)
log.Printf("%-s", resp)
log.Printf("%+s", resp)
```

## Setting
**NOTE** All settings methods is prefixed with Set
#### Set Timeout
``` go
req.Get("http://api.io").
	SetReadTimeout(40 * time.Second). // read timeout
	SetWriteTimeot(30 * time.Second). // write timeout
	SetDialTimeot(20 * time.Second).  // dial timeout
	SetTimeout(60 * time.Second).     // total timeout
	String()
```

#### Set Proxy
``` go
req.Get("http://api.io").
	SetProxy(func(r *http.Request) (*url.URL, error) {
		return url.Parse("http://localhost:40012")
	}).String()
```

#### Set Insecure TLS (Skip Verify Certificate Chain And Host Name)
``` go
req.Get("https://api.io").SetInsecureTLS(true).String()
```

TODO
