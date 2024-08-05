package myhttp

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReverseProxy(t *testing.T) {
	var (
		h1called  int
		h2called  int
		defcalled int
	)
	h1 := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		h1called++
	})
	h2 := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		h2called++
	})
	defh := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		defcalled++
	})

	var proxy ReverseProxyRouter
	proxy.Add(`^pero\.com$`, h1)
	proxy.Add(`^.+\.pero\.com$`, h2)
	proxy.Default(defh)

	r1, _ := http.NewRequest("GET", "https://pero.com/lala/pero", nil)
	r2, _ := http.NewRequest("GET", "https://oh.pero.com/lala/pero", nil)
	r3, _ := http.NewRequest("GET", "https://missed", nil)

	proxy.ServeHTTP(nil, r1)
	require.Equal(t, h1called, 1)
	require.Equal(t, h2called, 0)
	require.Equal(t, defcalled, 0)

	proxy.ServeHTTP(nil, r2)
	require.Equal(t, h1called, 1)
	require.Equal(t, h2called, 1)
	require.Equal(t, defcalled, 0)

	proxy.ServeHTTP(nil, r3)
	require.Equal(t, h1called, 1)
	require.Equal(t, h2called, 1)
	require.Equal(t, defcalled, 1)

	r4, _ := http.NewRequest("GET", "https://pero.com:8080/lala/pero", nil)

	proxy.ServeHTTP(nil, r4)
	require.Equal(t, h1called, 2)
	require.Equal(t, h2called, 1)
	require.Equal(t, defcalled, 1)

	r5, _ := http.NewRequest("GET", "/lala/pero", nil)
	r5.Header.Set("Host", "ribi.pero.com")

	proxy.ServeHTTP(nil, r5)
	require.Equal(t, h1called, 2)
	require.Equal(t, h2called, 2)
	require.Equal(t, defcalled, 1)
}
