# opentelemetry-jaeger-tracing

This is a runnable example of req, which uses the built-in tiny github sdk built on req to query and display the information of the specified user.

Best of all, it integrates seamlessly with jaeger tracing and is very easy to extend.

## How to run

First, use `docker` or `podman` to start a test jeager container (see jeager official doc: [ Getting Started](https://www.jaegertracing.io/docs/1.37/getting-started/#all-in-one)).

Then, run example:

```bash
go run .
```
```txt
Please give a github username: 
```

Input a github username, e.g. `imroc`:

```bash
$ go run .
Please give a github username: imroc
The moust popular repo of roc (https://imroc.cc) is req, which have 2500 stars
```

Then enter the Jaeger UI with browser (`http://127.0.0.1:16686/`), checkout the tracing details.

Run example again, try to input some username that doesn't exist, and check the error log in Jaeger UI.
