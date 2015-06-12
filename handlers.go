package advhttp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func NewPanicRecoveryHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			var err error
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("Unknown error")
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func NewDefaultCorsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ProcessCors(w, r)
		h.ServeHTTP(w, r)
	})
}

func NewCorsHandler(h http.Handler, cors *Cors) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cors.ProcessCors(w, r)
		h.ServeHTTP(w, r)
	})
}

func NewForwardedLoggingHandler(h http.Handler, log io.Writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("X-User-Id")
		r.Header.Del("X-Client-Id")
		trw := NewResponseWriter(w)
		origURI := r.URL.RequestURI()
		h.ServeHTTP(trw, r)
		r.URL, _ = url.Parse(origURI)
		fmt.Fprintln(log, trw.LogCommonExtendedForwarded(r))
	})
}

func NewLoggingHandler(h http.Handler, log io.Writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("X-Forwarded-For")
		r.Header.Del("X-User-Id")
		r.Header.Del("X-Client-Id")
		trw := NewResponseWriter(w)
		origURI := r.URL.RequestURI()
		h.ServeHTTP(trw, r)
		r.URL, _ = url.Parse(origURI)
		fmt.Fprintln(log, trw.LogCommonExtended(r))
	})
}

func NewReverseProxyHandler(destinationURL *url.URL, stripListenPath bool, listenPath string) http.Handler {
	return NewGatewayReverseProxy(destinationURL, stripListenPath, listenPath)
}
