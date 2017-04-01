req
==============
req is a super light weight and super easy-to-use golang http request library.

# Document
[中文](README_EN.md)


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

## Set Body, Params, Headers
#### Body
``` go
r := req.Post(url).Body(`hello req`)
r.GetBody() // hello req

r.BodyJson(&struct { // it could also could be string or []byte
	Usename  string `json:"usename"`
	Password string `json:"password"`
}{
	Username: "req",
	Password: "req",
})
r.GetBody() // {"username":"req","password","req"}

r.BodyXml(&foo)
```

#### Params
**note** it will url encode your params automatically.
``` go
r := req.Get("http://api.foo").Params(req.M{
	"username": "req",
	"password": "req",
})
r.GetUrl() // http://api.foo?username=req&password=req

r = req.Post(url).Param("username", "req")
r.GetBody() // username=req
```

#### Headers
``` go
r := req.Get("https://api.foo/get")
r.Headers(req.M{
	"Referer":    "http://api.foo",
	"User-Agent": "Chrome/57.0.2987.110",
})
/*
	GET https://api.foo/get HTTP/1.1
	Referer:http://api.foo
	User-Agent:Chrome/57.0.2987.110
*/
fmt.Printf("%+r", r)
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

## Print Detail
Sometimes you might want to dump the detail about the http request and response for debug or logging reasons. 
There are several format to print these detail infomation.


#### Default Print
Use `%v` or `%s` to get the info in default format.
``` go
r := req.Get("http://api.foo/get")
log.Printf("%v", r) // GET http://api.foo/get {"success":true,"data":"hello req"}
r = req.Post("http://api.foo/post").Body(`{"uid":"1"}`)
log.Println(r) // POST http://api.foo/post {"uid":"1"} {"success":true,"data":{"name":"req"}}
```
**NOTE** it will add newline if possible, keep it looks pretty. 


#### Print All Infomation
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
As you can see, it will print the request Method,URL,Proto,[Request Header],[Request Body],[Response Header],[Response Body]


#### Print In Oneline
Use `%-v` or `%-s` keeps info in one line (delete all blank characters if possible), this is useful while logging.
``` go
r := req.Get("http://api.foo/get")
// it print every thing in one line, even if '\n' exsist in reqeust body or response body.
log.Printf("%-v\n",r) // GET http://api.foo/get {"code":3019,"msg":"system busy"}
```


#### Print Request Only (No Response Info)
Use `%r`, `%+r` or `%-r` only print the request itself, no response.
``` go
r := req.Post("https://api.foo").Body(`name=req`)
fmt.Printf("%r", r) // POST https://api.foo name=req
```
**NOTE** in other format above, it will execute the underlying request to get response if the request is not executed yet, you can disable that by using this format.


#### Print Response Only
You need get the *req.Response, use `%v`,`%s`,`%+v`,`%+s`,`%-v`,`%-s` to print formatted response info.
``` go
resp := req.Get(url).Response()
log.Println(resp)
log.Printf("%-s", resp)
log.Printf("%+s", resp)
```

## Setting
**NOTE** All settings methods is prefixed with Set or Enable
#### Set Timeout
``` go
req.Get(url).
	SetReadTimeout(40 * time.Second). // read timeout
	SetWriteTimeout(30 * time.Second). // write timeout
	SetDialTimeout(20 * time.Second).  // dial timeout
	SetTimeout(60 * time.Second).     // total timeout
	String()
```

#### Set Proxy
``` go
req.Get(url).
	SetProxy(func(r *http.Request) (*url.URL, error) {
		return url.Parse("http://localhost:40012")
	}).String()
```

#### Allow Insecure Https (Skip Verify Certificate Chain And Host Name)
``` go
req.Get(url).EnableInsecureTLS(true).String()
```

#### Reuse Setting
if you care about performance very much, you can reuse the setting. (the internal `http.Client` will be created only once)

create a Setting:
``` go
setting := &req.Setting{
	InsecureTLS: true,
	Timeout:     20 * time.Second,
}
```
this is same as:
``` go
setting := req.New().SetTimeout(20 * time.Second).EnableInsecureTLS(true).GetSetting()
```
then call Setting method to set the settings:
``` go
req.Get(url).Setting(setting).Bytes()
```

#### More Setting
req uses `http.Client` and `http.Transport` internally, and you can easily modify it, making it has much more potential. You can call `GetClient` or `GetTransport` to get the generated `*http.Client` and `*http.Transport`
``` go
setting := &req.Setting{
	InsecureTLS: true,
	Timeout:     20 * time.Second,
}
setting.GetTransport().MaxIdleConns = 100
setting.GetClient().Jar, _ = cookiejar.New(nil) // manage cookie
req.Get(url).Setting(setting).Bytes()
```