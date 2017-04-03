req
==============
[![GoDoc](https://godoc.org/github.com/imroc/req?status.svg)](https://godoc.org/github.com/imroc/req)

req 是一个非常轻量级、简单易用的go语言http请求库


# 快速开始
## 安装
``` sh
go get github.com/imroc/req
```

## 基础用法 
``` go
req.Get(url).String() // 获取响应返回string
req.Post(url).Body(body).ToJSON(&foo) // 设置请求体(可以是string或[]byte)，获取响应并将相应体转成struct
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

## 设置 请求体, 请求参数, 请求头
#### 请求体
``` go
r := req.Post(url).Body(`hello req`)
r.GetBody() // hello req

r.BodyJSON(&struct { // 也可以是string或[]byte
	Usename  string `json:"usename"`
	Password string `json:"password"`
}{
	Username: "req",
	Password: "req",
})
r.GetBody() // {"username":"req","password","req"}

r.BodyXML(&foo)
```

#### 请求参数
**注意** 会默认自动加Content-Type的请求头
``` go
r := req.Get("http://api.foo").Params(req.M{
	"username": "req",
	"password": "req",
})
r.GetUrl() // http://api.foo?username=req&password=req

r = req.Post(url).Param("username", "req")
r.GetBody() // username=req
```

#### 请求头
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
fmt.Printf("%+r", r) // 不懂"%r"? 别急，后面会讲
```

## 获取响应
```go
r := req.Get(url)
r.Response()   // *req.Response
r.String()     // string
r.Bytes()      // []byte
r.ToJSON(&foo) // json->struct
r.ToXML(&bar)  // xml->struct

// Receive开头的函数，如果请求出错会返回错误
_, err = r.ReceiveResponse()
_, err = r.ReceiveString()
_, err = r.ReceiveBytes()
```
**注意:** 当你调用上面的方法获取响应的时候，底层的请求默认只会发起一次，响应结果缓存在req.Request内部。如果想要再次发起请求，可以调用`Do`强制发起请求。或者调用`Undo`，当下次获取响应的时候不再返回缓存，而是真正发起请求并返回响应。

## 打印详情
再调试接口或打印日志的时候我们可能需要输出请求的一些详细信息，req提供了几种格式帮助你打印请求详情


#### 默认格式
使用 `%v` 或 `%s` 格式打印默认格式的请求详情
``` go
r := req.Get("http://api.foo/get")
log.Printf("%v", r) // GET http://api.foo/get {"success":true,"data":"hello req"}
r = req.Post("http://api.foo/post").Body(`{"uid":"1"}`)
log.Println(r) // POST http://api.foo/post {"uid":"1"} {"success":true,"data":{"name":"req"}}
```
**注意** 为了让输出好看，有时会智能的新增空行。


#### 打印所有信息
使用 `%+v` 或 `%+s` 输出所有请求相关的信息
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
从上面可以看到，`%+v`格式输出会尽可能打印所有信息，包括请求方法、请求地址、请求协议版本、请求头、请求体、响应头、响应体

#### 打印在一行
使用 `%-v` 或 `%-s` 格式只输出必要的信息并保持信息打印在一行(删除请求体和响应体所有空白字符)，这在日志记录的时候非常有用
``` go
r := req.Get("http://api.foo/get")
log.Printf("%-v\n",r) // GET http://api.foo/get {"code":3019,"msg":"system busy"}
```


#### 只打印请求本身(不要响应)
使用 `%r`、 `%+r` 或 `%-r` 格式。
``` go
r := req.Post("https://api.foo").Body(`name=req`)
fmt.Printf("%r", r) // POST https://api.foo name=req
```
**注意** 使用之前其它打印请求的格式，因为要输出响应，所以如果内部的请求还没有发起的话，它会先发起请求并获取响应。如果你只想打印请求本生的话，可以使用上面这种格式，它不会发起请求


#### 只打印响应
如果只要响应信息，你需要调用`Response`或`ReceiveResponse`获取响应，返回`*req.Response`，然后用 `%v`、`%s`、`%+v`、`%+s`、`%-v`或`%-s`格式输出想要的信息
``` go
resp := req.Get(url).Response()
log.Println(resp)
log.Printf("%-s", resp)
log.Printf("%+s", resp)
```

## 设置
#### 设置超时限制
``` go
req.Get(url).
	Timeout(60 * time.Second).             // 总超时时间限制
	TimeoutRead(40 * time.Second).         // 读取超时
	TimeoutWrite(30 * time.Second).        // 写入超时
	TimeoutDial(20 * time.Second).         // 建立连接超时
	TimeoutTLSHandshake(10 * time.Second). // https握手超时
	String()
```

#### 设置代理
``` go
req.Get(url).
	Proxy(func(r *http.Request) (*url.URL, error) {
		return url.Parse("http://localhost:40012")
	}).String()
```

#### 允许不安全的https(忽略校验证书和域名)
``` go
req.Get(url).InsecureTLS(true).String()
```


#### 更多设置
req 内部使用标准库的`http.Client`和`http.Transport`，你可以获取出来任意进行修改，非常灵活。调用`GetClient`和 `GetTransport`分别可以获取根据设置生成的`*http.Client`和`*http.Transport`。
``` go
r := req.Get(url)
r.GetTransport().MaxIdleConns = 100
r.GetClient().Jar, _ = cookiejar.New(nil) // 管理cookie
```


## 不同请求共享相同属性
`Merge` 函数可以将另外一个请求的一些属性合并进当前的请求
``` go
// 创建需要共享的属性
attr := req.New()
attr.Header("User-Agent", "V1.1.1")
attr.Timeout(10 * time.Second)
attr.Param("access_token", token)
attr.Proto("HTTP/2")
attr.Cookie(&http.Cookie{Name: "sessionid", Value: "FHJ67GHJ67G8G65HJJ", Path: "/", MaxAge: 86400})
attr.InsecureTLS(true)

// 合并属性并发起请求
req.Get(api1).Merge(attr).String()
req.Get(api2).Merge(attr).InsecureTLS(false).String()
```