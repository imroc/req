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
