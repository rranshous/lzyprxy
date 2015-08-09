package main

import (
	"./atc"
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

	atc := atc.NewAirTrafficControl()

	type responseStatus struct {
		host   string
		status int
		length int64
	}

	type requestStatus struct {
		host string
	}

	requests := make(chan requestStatus)
	responses := make(chan responseStatus)

	go func() {
		for {
			msg := <-requests
			fmt.Println("request host", msg.host)
		}
	}()

	go func() {
		for {
			msg := <-responses
			fmt.Println("respons host", msg.host, "status", msg.status, "len", msg.length)
			// report our status to air traffic control
			if msg.status > 499 {
				atc.ReportFailure(msg.host)
			} else {
				atc.ReportSuccess(msg.host)
			}
		}
	}()

	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			// todo: fail request if not cleared
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
