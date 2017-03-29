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
Sometimes you may want to output the detail about the http request for debug or logging reasons. 
There are several ways to output the detail.
### Default Format
By default, the output format is: Method URL \[Reqeust Body\] \[Response Body\]
``` go
r := req.Get("http://api.xxx.com/get")
fmt.Println(r) // GET http://api.xxx.com/get {"success":true,"data":"hello roc"}
r = req.Post("http://api.xxx.com/post").Body(`{"uid":"1"}`)
fmt.Println(r) // // POST http://api.xxx.com/post {"uid":"1"} {"success":true,"data":{"name":"roc"}}
```
NOTE: it will execute the reqeust to get response if the reqeust is not executed(disable this by use %r,%+r or %-r format, only output the request itself), and it will add newline if possible, keep the output looks pretty.
### Maximum Infomation Format
Use %+v or %+s to get the maximum detail infomation.
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
	Etag:"2bd1-52ce6b6c4bc00"
	Server:bfe/1.0.8.18
	Set-Cookie:BAIDUID=EEED32735E8B09B97A3D3871392F82C5:FG=1; expires=Thu, 29-Mar-18 09:18:13 GMT; max-age=31536000; path=/; domain=.baidu.com; version=1
	Set-Cookie:__bsi=3154899087195606076_00_0_I_R_2_0303_C02F_N_I_I_0; expires=Wed, 29-Mar-17 09:18:18 GMT; domain=www.baidu.com; path=/
	Expires:Thu, 30 Mar 2017 09:18:13 GMT
	Last-Modified:Mon, 29 Feb 2016 11:11:44 GMT
	Cache-Control:max-age=86400
	Date:Wed, 29 Mar 2017 09:18:13 GMT
	Connection:keep-alive
	P3p:CP=" OTI DSP COR IVA OUR IND COM "
	Accept-Ranges:bytes
	Vary:Accept-Encoding,User-Agent
	Content-Type:text/html

	{"success":true,"data":{"name":"roc"}}
*/
fmt.Printf("%+v\n", r)
```
As you can see, it will output the request Method,URL,Proto,[Request Header],[Request Body],[Response Header],[Response Body]
### One Line Format
Sometimes you want to keep all of infomation in one line(delete all blank character), it is useful when logging(you can easily find the infomation use the cammand like grep). Just use %-v or %-s.
``` go
r := req.Get("http://api.xxx.com/get")
// it output every thing in one line, even if '\n' exsist in reqeust body or response body.
fmt.Printf("%-v\n",r) // GET http://api.xxx.com/get {"success":true,"data":"hello roc"}
```
