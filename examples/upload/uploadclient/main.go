package main

import "github.com/imroc/req/v3"

func main() {
	req.EnableDumpNoRequestBody()
	req.SetFile("files", "../../../README.md").
		SetFile("files", "../../../LICENSE").
		SetFormData(map[string]string{
			"name":  "imroc",
			"email": "roc@imroc.cc",
		}).
		Post("http://127.0.0.1:8888/upload")
	/* Output
	POST /upload HTTP/1.1
	Host: 127.0.0.1:8888
	User-Agent: req/v2 (https://github.com/imroc/req)
	Transfer-Encoding: chunked
	Content-Type: multipart/form-data; boundary=6af1b071a682709355cf5fb15b9cf9e793df7a45e5cd1eb7c413f2e72bf6
	Accept-Encoding: gzip

	HTTP/1.1 200 OK
	Content-Type: text/plain; charset=utf-8
	Date: Tue, 25 Jan 2022 09:40:36 GMT
	Content-Length: 76

	Uploaded successfully 2 files with fields name=imroc and email=roc@imroc.cc.
	*/
}
