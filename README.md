# httpgo 
[![Build Status](https://travis-ci.org/p4ali/httpgo.svg?branch=master)](https://travis-ci.org/p4ali/httpgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/p4ali/httpgo)](https://goreportcard.com/report/github.com/p4ali/httpgo)
[![Documentation](https://godoc.org/github.com/p4ali/httpgo?status.svg)](http://godoc.org/github.com/p4ali/httpgo)
[![Coverage Status](https://coveralls.io/repos/github/p4ali/httpgo/badge.svg?branch=master)](https://coveralls.io/github/p4ali/httpgo?branch=master)
[![GitHub issues](https://img.shields.io/github/issues/p4ali/httpgo.svg)](https://github.com/p4ali/httpgo/issues)
[![license](https://img.shields.io/github/license/p4ali/httpgo.svg?maxAge=2592000)](https://github.com/p4ali/httpgo/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/p4ali/httpgo.svg?label=Release)](https://github.com/p4ali/httpgo/releases)

## Why another image for http server

When we proto-type or debug in container runtime, we often need a smaller http server image (for quick download)
equipped with necessary network utils like [busybox](https://hub.docker.com/_/busybox).
 
Existent http server such as [httpin](https://httpbin.org/), normally has a large image more than 100M.

To get similar function, but with a small size, `httpgo` is writen with golang, and the total image size is 
less than 20M together with busybox tools.  

## Http Server with Golang

A HTTP server with health check endpoint. By manipulate the `/health` endpoint, you can return `503` for all
endpoints (except the `POST /health` of course) when `/health` unhealthy, or otherwise behave normally.

## Headers in Response

**The following headers are returned in each call**

|Header                       | Value                                                    |
|:----------------------------|:---------------------------------------------------------|
|Access-Control-Allow-Origin  | `*`                                                      | 
|Access-Control-Allow-Headers | `Content-Range, Content-Disposition, Content-Type, ETag` |
|echo-server-ip               | server ip, e.g., `1.2.3.4`                               |
|echo-server-version          | server version, e.g., `0.0.1`                            |
|echo-server-name             | server name, e.g., `httpgo`                              |

## Endpoints

**Endpoints**

|Endpoint             |Method | Description                                                                        |
|:--------------------|:------|:-----------------------------------------------------------------------------------|
| /callother          |POST   | call other urls in `\r\n` separated string in request body, e.g., `url1\r\nurl2`   |
| /debug              |GET    | return server info and env                                                         |
| /delay/{x}          |GET    | return 200 after delay x milliseconds                                              |
| /echo/{msg}         |GET    | return 200 and print msg                                                           |
| /health             |GET    | return health setting                                                              |
| /health             |POST   | update the health setting, e.g., /health?value=false                               |
| /health             |HEAD   | return health setting                                                              |
| /status/{code}      |GET    | return given `code` as status                                                      |


## Development

```$bash
make
make test
```

We currently use `go mod`, which will download dependencies in CI. But if for any reason your CI machine
can not download the dependencies in your/company network, you can vendor the dependencies:
```
go mod vendor
```
which will download the dependencies and put into `vendor` folder.

**NOTE**: you need set `GO111MODULE=off` when using `vendor` folder, otherwise, `go mod` will still try 
download dependencies..

If you want to clean up dependencies, run `go mod tidy`.


## Published docker image

The image is published as `p4ali/httpgo` on [docker hub](https://hub.docker.com/r/p4ali/httpgo).
The tag containing  `nonroot` is the the image built from `nonroot` branch, where user can only run 
as non-root. Other tag is for image built from master where user can run as root.

## Publish binary

```$bash
make version=1.0.0 goos=windows release
make version=1.0.0 goos=linux release
make version=1.0.0 goos=darwin release  # default make release
```

After that, you can use `brew install p4ali/tools/httpgo` to install `httgo` to `/usr/local/Cellar/httpgo/1.0.0` 
and create a symbol link at `/usr/local/bin/httpgo`.