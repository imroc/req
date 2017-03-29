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
fmt.Println(r) // GET http://api.xxx.com/get {"success":true,"data":"hello roc"}
```

## POST
``` go
r := req.Post("http://api.xxx.com/post").Body(`{"uid":"1"}`)
fmt.Printf("resp:%s\n",r.String()) // resp:{"success":true,"data":{"name":"roc"}}
fmt.Println(r) // POST http://api.xxx.com/post {"uid":"1"} {"success":true,"data":{"name":"roc"}}
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