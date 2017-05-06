# req
A golang http request library.


Features
========

- light weight
- simple
- useful optional params (header,param,file...)
- easy result converting (ToXML,ToJSON,ToFile)
- easy debugging and logging (print info)

Install
=======
``` sh
go get github.com/imroc/req
```

Quick Start
=======
request = method + url [+ options]  
the options can be headers, params, files or body etc
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
and the `%-v` format is similar to `%v`, the only difference is that it removes all blank characters, keep content minimal and in one line, it is useful while logging.

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
