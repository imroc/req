req
==============
req is a super light weight and super easy-to-use  http request library.


# Quick Start
## Install
``` sh
go get github.com/imroc/req
```


## GET
``` go
r := req.Get("http://api.xxx.com/get")
fmt.Printf("resp:%s\n",r.String()) // resp:{"success":true,"data":"hello roc"}
```


## POST
``` go
r := req.Post("http://api.xxx.com/post").Body(`{"uid":"1"}`)
fmt.Printf("resp:%s\n",r.String()) // resp:{"success":true,"data":{"name":"roc"}}
```


## Param
``` go
// GET params
r := req.Get("http://api.xxx.com").Param("p1", "1").Params(req.M{
	"p2": "2",
	"p3": "3",
})
fmt.Println(r.GetUrl()) // http://api.xxx.com?p1=1&p2=2&p3=3

// POST params
r = req.Post("http://api.xxx.com").Params(req.M{
	"p1": "1",
	"p2": "2",
})
fmt.Println(string(r.GetBody())) // p1=1&p2=2
```


## Header
``` go
r := req.Get("http://api.xxx.com/something").Headers(req.M{
	"Referer":    "http://api.xxx.com",
	"User-Agent": "custom client",
})
```

## Timeout
TODO


## Response
```go
r := req.Get("http://api.xxx.com")
resp := r.Response() // *req.Response
str := r.String() // string
bs := r.Bytes()   // []byte
var result struct {
	Success bool   `json:"success" xml:"success"`
	Data    string `json:"data" xml:"data"`
}
r.ToJson(&result) // json
//r.ToXml(&result)  // xml
```

## Error
This library will never panic, it returns error when you need to handle it. 
When you call methods like r.String() or r.Bytes(), it will ignore the error and return zero value if error happens. 
And when you call methods like r.ReceiveString() which start with Receive it will return error if error happens.

```go
r := req.Get("http://api.xxx.com")
resp, err := r.ReceiveResponse()
if err != nil {
	fmt.Println("error:", err)
	return
}
fmt.Println("resp:", resp.String())
```

## Formatted Infomation
Sometimes you might want to output the detail about the http request for debug or logging reasons. 
There are several ways to output these detail infomation.


#### Default Format
Use `%v` or `%s` to get the default infomation.
``` go
r := req.Get("http://api.xxx.com/get")
fmt.Printf("%v", r) // GET http://api.xxx.com/get {"success":true,"data":"hello roc"}
r = req.Post("http://api.xxx.com/post").Body(`{"uid":"1"}`)
fmt.Println(r) // POST http://api.xxx.com/post {"uid":"1"} {"success":true,"data":{"name":"roc"}}
```
By default, the output format is: Method URL \[Reqeust Body\] \[Response Body\] and it will add newline if possible, keep the output looks pretty. 


#### Maximal Format
Use `%+v` or %+s to get the maximal detail infomation.
``` go
r := req.Post("http://api.xxx.com/post").Headers(req.M{
	"Referer": "http://api.xxx.com",
}).Params(req.M{
	"p1": "1",
	"p2": "2",
})
/*
	POST http://api.xxx.com/post HTTP/1.1

	Referer:http://api.xxx.com
	Content-Type:application/x-www-form-urlencoded

	p1=1&p2=2

	HTTP/1.1 200 OK
	Server:nginx
	Set-Cookie:bla=3154899087195606076; expires=Wed, 29-Mar-17 09:18:18 GMT; domain=api.xxx.com; path=/
	Expires:Thu, 30 Mar 2017 09:18:13 GMT
	Cache-Control:max-age=86400
	Date:Wed, 29 Mar 2017 09:18:13 GMT
	Connection:keep-alive
	Accept-Ranges:bytes
	Content-Type:application/json

	{"success":true,"data":{"name":"roc"}}
*/
fmt.Printf("%+v\n", r)
```
As you can see, it will output the request Method,URL,Proto,[Request Header],[Request Body],[Response Header],[Response Body]


#### Minimal Format
Sometimes you might want to keep all infomation in one line (delete all blank character if possible), it is useful while logging (you can easily find the infomation using the cammand like grep). Try `%-v` or `%-s`.
``` go
r := req.Get("http://api.xxx.com/get")
// it output every thing in one line, even if '\n' exsist in reqeust body or response body.
log.Printf("%-v\n",r) // GET http://api.xxx.com/get {"success":false,"msg":"system busy"}
```
now if you want to find out which ones is not success of that particular url, and because it's one line per request, so you can use cammand like `more log.log | grep "http://api.xxx.com/get" | grep "\"success\":false"` to get the answer.


#### Request Only Format (No Response Info)
It will execute the request to get response if the request is not executed yet by default, you can disable this by using `%r`, `%+r` or `%-r` format, only output the request itself.
``` go
r := req.Get("https://www.baidu.com")
fmt.Printf("%r", r) // GET https://www.baidu.com HTTP/1.1
```


## Reuse Request
By default, when calling methods like `r.String()` `r.ReceiveBytes()` to get the result, it will execute the request if it's not been executed yet, and do not execute it next time. But, sometings we need to reuse the request, maybe just retry, maybe change a param and retry.

For example, some api need access token, and the token is expired, if you call that api you will got error, then you need to refresh token and try again.
``` go
r := req.Get("http://api.xxx.com").Param("access_token", token)
fmt.Println(r) // GET http://api.xxx.com?access_token=HJKJ354HK67FGHJ75 {"errcode":42001,"errmsg":"access token expired"}
token = RefreshToken()
r.Param("access_token", token)
fmt.Println(r) // GET http://api.xxx.com?access_token=G7GJ6DFH546H86F6G {"errcode":0,"errmsg":"OK"}
```

## Proxy
TODO