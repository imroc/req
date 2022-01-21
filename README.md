# req

[![GoDoc](https://pkg.go.dev/badge/github.com/imroc/req.svg)](https://pkg.go.dev/github.com/imroc/req)

A golang http request library for humans.

## Features

* Simple and chainable methods for client and request settings.
* Rich syntax sugar, greatly improving development efficiency.
* Automatically detect charset and decode to utf-8.
* Powerful debugging capabilities (logging, tracing, and event dump the requests and responses content).
* The settings can be dynamically adjusted, making it possible to debug in the production environment.
* Easy to integrate with existing code, just replace client's Transport you can dump requests and reponses to debug.

## Install

``` sh
go get github.com/imroc/req/v2@v2.0.0-alpha.0
```

## Usage

Import req in your code:

```go
import "github.com/imroc/req/v2"
```

Prepare client:

```go
req.C().UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:95.0) Gecko/20100101 Firefox/95.0")
```
