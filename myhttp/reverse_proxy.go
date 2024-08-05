package myhttp

import (
	"net/http"
	"regexp"
	"strings"
)

type reverseProxyData struct {
	r *regexp.Regexp
	h http.Handler
}

type ReverseProxyRouter struct {
	routes     []reverseProxyData
	defhandler http.Handler
}

func (router *ReverseProxyRouter) Add(pattern string, handler http.Handler) {
	if router.routes == nil {
		router.routes = []reverseProxyData{}
	}

	router.routes = append(router.routes, reverseProxyData{
		r: regexp.MustCompile(pattern),
		h: handler,
	})
}

func (router *ReverseProxyRouter) Default(handler http.Handler) {
	router.defhandler = handler
}

func (router *ReverseProxyRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Hostname()
	if host == "" {
		host = r.Header.Get("host")
	}
	if host == "" {
		host = r.Host
	}

	idx := strings.LastIndex(host, ":")
	if idx > 0 {
		host = host[:idx]
	}

	for _, route := range router.routes {
		if route.r.MatchString(host) {
			route.h.ServeHTTP(w, r)
			return
		}
	}

	if router.defhandler == nil {
		panic("no default handler")
	}

	router.defhandler.ServeHTTP(w, r)
}
