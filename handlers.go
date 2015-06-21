package advhttp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// NewPanicRecoveryHandler will wrap a handler in a recover function that will
// catch any panics that occur, and gracefully (actually return a response) handle
// the panic by returning a 500 Internal Server Error response with the panic
// error as the body.
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

// Returns a handler that wraps the given handler with Cross Origin Resource
// Sharing response headers. It uses the default settings which are very
// permissive. The settings can be changed directly on the default cors object,
// or alternatively you can create your own cors object and use `NewCorsHandler()`
func NewDefaultCorsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ProcessCors(w, r)
		h.ServeHTTP(w, r)
	})
}

// Returns a handler with a custom cors object and uses that methods `ProcessCors()`
// function before calling the wrapped handler.
func NewCorsHandler(h http.Handler, cors *Cors) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cors.ProcessCors(w, r)
		h.ServeHTTP(w, r)
	})
}

// Returns a logging handler that wraps the given handler, and logs output to the
// given io.Writer. The logging format is a variation of the `Common Log Format`.
// The Forwarded Variant will utilize the `X-Forwarded-*` headers to log ip, host,
// and proto.
func NewForwardedLoggingHandler(h http.Handler, log io.Writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("X-User-Id")
		r.Header.Del("X-Client-Id")
		trw := NewResponseWriter(w)
		origURI := r.URL.RequestURI()
		start := time.Now()
		h.ServeHTTP(trw, r)
		r.URL, _ = url.Parse(origURI)
		fmt.Fprintln(log, trw.LogWithOptions(r, true, time.Now().Sub(start)))
	})
}

// Returns a logging handler that wraps the given handler, and logs output to the
// given io.Writer. The logging format is a variation of the `Common Log Format`.
func NewLoggingHandler(h http.Handler, log io.Writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("X-User-Id")
		r.Header.Del("X-Client-Id")
		trw := NewResponseWriter(w)
		origURI := r.URL.RequestURI()
		start := time.Now()
		h.ServeHTTP(trw, r)
		r.URL, _ = url.Parse(origURI)
		fmt.Fprintln(log, trw.LogWithOptions(r, true, time.Now().Sub(start)))
	})
}

// Returns a reverse proxy handler that wraps the GatewayReverseProxy structure and
// functions to provide api gateway functionality.
func NewReverseProxyHandler(proxyName string, destinationURL *url.URL, stripListenPath bool, listenPath string) http.Handler {
	return NewGatewayReverseProxy(destinationURL, stripListenPath, listenPath)
}
