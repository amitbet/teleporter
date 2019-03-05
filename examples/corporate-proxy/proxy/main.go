package main

import (
    "github.com/elazarl/goproxy"
    "log"
    "net/http"
	"fmt"
)

func main() {
    proxy := goproxy.NewProxyHttpServer()
    proxy.Verbose = true
	proxy.OnRequest().HandleConnect(goproxy.FuncHttpsHandler(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		fmt.Printf("Got CONNECT Host: %v, URL: %v ReqHost: %v\n", host, ctx.Req.URL.String(), ctx.Req.Host)
		return goproxy.OkConnect, host
	}))
    log.Fatal(http.ListenAndServe(":8080", proxy))
}
