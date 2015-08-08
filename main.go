package main

import (
	"flag"
	"fmt"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
)

func main() {
	verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr := flag.String("addr", ":8080", "proxy listen address")
	flag.Parse()
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verbose

	type responseStatus struct {
		host   string
		status int
		length int64
	}

	type requestStatus struct {
		host string
	}

	requests := make(chan requestStatus)
	requestCounts := make(map[string]int64)
	responses := make(chan responseStatus)
	responseCounts := make(map[string]int64)

	go func() {
		for {
			msg := <-requests
			fmt.Println("request host", msg.host)
			requestCounts[msg.host] += 1
			for host, count := range requestCounts {
				fmt.Println("Rq[", host, "]", count)
			}
		}
	}()

	go func() {
		for {
			msg := <-responses
			fmt.Println("response host", msg.host, "status", msg.status, "len", msg.length)
			responseCounts[msg.host] += 1
			for host, count := range responseCounts {
				fmt.Println("Re[", host, "]", count)
			}
		}
	}()

	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			requests <- requestStatus{r.Host}
			return r, nil
		})

	proxy.OnResponse().DoFunc(
		func(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			if r != nil {
				responses <- responseStatus{r.Request.Host, r.StatusCode, r.ContentLength}
			}
			return r
		})

	fmt.Println("STARTING")
	log.Fatal(http.ListenAndServe(*addr, proxy))
}
