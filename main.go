package main

import (
	"github.com/elazarl/goproxy"
	"log"
	"flag"
	"net/http"
	"fmt"
)

func main() {
	verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr := flag.String("addr", ":8080", "proxy listen address")
	flag.Parse()
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verbose

	type responseStatus struct {
		host string
		status int
		length int64
	}

	responses := make(chan responseStatus)
	go func() {
		for {
			msg := <-responses;
			fmt.Println("host",msg.host,"status",msg.status,"len",msg.length)
		}
	}()

	proxy.OnRequest().DoFunc(
		func(r *http.Response, ctx *goproxy.ProxyCtx)(*http.Response) {
			responses<-responseStatus{r.Request.Host, r.StatusCode, r.ContentLength}
			return r,nil
	})
	
	/*
	proxy.OnRequest().DoFunc(
		func(r *http.Request,ctx *goproxy.ProxyCtx)(*http.Request,*http.Response) {
			responses<-responseStatus{r.Host, 200}
			return r,nil
	})
	*/
	fmt.Println("STARTING")
	log.Fatal(http.ListenAndServe(*addr, proxy))
}
