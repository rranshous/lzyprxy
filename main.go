package main

import (
	"./atc"
	"flag"
	"fmt"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
)

func isFailure(status int) bool {
	return status > 499 || status == -1
}

type responseStatus struct {
	host   string
	status int
	length int64
}

type requestStatus struct {
	host string
}

func main() {
	verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr := flag.String("addr", ":8080", "proxy listen address")
	flag.Parse()
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verbose

	atc := atc.NewAirTrafficControl()

	responses := make(chan responseStatus)

	go func() {
		for {
			msg := <-responses
			fmt.Println("respons host", msg.host, "status", msg.status, "len", msg.length)

			if isFailure(msg.status) {
				atc.ReportFailure(msg.host)
			} else {
				atc.ReportSuccess(msg.host)
			}
		}
	}()

	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			cleared := atc.GetClearance(r.Host)

			if !cleared {
				fmt.Println("NOT CLEARED!")
				return r, goproxy.NewResponse(r,
					goproxy.ContentTypeText, http.StatusGatewayTimeout,
					"circuit broken")
			}

			return r, nil
		})

	proxy.OnResponse().DoFunc(
		func(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			if r != nil {
				responses <- responseStatus{r.Request.Host, r.StatusCode, r.ContentLength}
			} else {
				fmt.Println("response ERROR:", ctx.Error)
				responses <- responseStatus{ctx.Req.Host, -1, -1}
			}
			return r
		})

	fmt.Println("STARTING")
	log.Fatal(http.ListenAndServe(*addr, proxy))
}
