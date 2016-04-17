package advhttp

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"path"
	"strings"
)

// The util response writer gives you the ability to hook into the response
// writer lifecycle. This Includes being able to have a callback to re-write
// the headers before the body is written out. This is mostly useful for a
// response writer to give to a reverse proxy. You can then re-write the
// location header as well as add in via headers, etc.
type UtilResponseWriter struct {
	w                   http.ResponseWriter
	sentHeaders         bool
	SendHeadersCallback func(http.Header)
}

// Creates a new UtilResponseWriter wrapping the given http.UtilResponseWriter
func NewUtilResponseWriter(w http.ResponseWriter, hcb func(http.Header)) *UtilResponseWriter {
	return &UtilResponseWriter{w: w, SendHeadersCallback: hcb}
}

// Creates a new UtilResponseWriter wrapping the given http.ResponseWriter with
// a location rewriting and via header adding response writer
func NewRPResponseWriter(w http.ResponseWriter, prefix string, via string) *UtilResponseWriter {
	return &UtilResponseWriter{w: w, SendHeadersCallback: NewReverseProxyHeadersCallback(prefix, via)}
}

func NewRewriteLocationHeaderCallback(prefix string) func(http.Header) {
	return func(headers http.Header) {
		oldLocation := headers.Get("Location")
		if oldLocation != "" {
			if strings.HasPrefix(oldLocation, "/") {
				if strings.HasSuffix(oldLocation, "/") {
					headers.Set("Location", path.Join(prefix, oldLocation)+"/")
				} else {
					headers.Set("Location", path.Join(prefix, oldLocation))
				}
			}
		}
	}
}

func NewReverseProxyHeadersCallback(prefix string, via string) func(http.Header) {
	return func(headers http.Header) {
		oldLocation := headers.Get("Location")
		if oldLocation != "" {
			if strings.HasPrefix(oldLocation, "/") {
				if strings.HasSuffix(oldLocation, "/") {
					headers.Set("Location", path.Join(prefix, oldLocation)+"/")
				} else {
					headers.Set("Location", path.Join(prefix, oldLocation))
				}
			}
		}
		if headers.Get("Via") == "" {
			headers.Set("Via", via)
		} else {
			headers.Set("Via", via+", "+headers.Get("Via"))
		}
	}
}

func (urw *UtilResponseWriter) Header() http.Header {
	return urw.w.Header()
}

func (urw *UtilResponseWriter) WriteHeader(status int) {
	urw.SendHeadersCallback(urw.Header())
	urw.sentHeaders = true
	urw.w.WriteHeader(status)
}

func (urw *UtilResponseWriter) Write(bytes []byte) (int, error) {
	if !urw.sentHeaders {
		urw.SendHeadersCallback(urw.Header())
		urw.sentHeaders = true
	}
	return urw.w.Write(bytes)
}

func (urw *UtilResponseWriter) GetFlusher() (flusher http.Flusher, ok bool) {
	flusher, ok = urw.w.(http.Flusher)
	return
}

func (urw *UtilResponseWriter) Flush() {
	if flusher, ok := urw.w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (urw *UtilResponseWriter) GetHijacker() (hijacker http.Hijacker, ok bool) {
	hijacker, ok = urw.w.(http.Hijacker)
	return
}

func (urw *UtilResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := urw.w.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.New("Couldn't cast responsewriter to hijacker")
}

func (urw *UtilResponseWriter) GetCloseNotifier() (closeNotifier http.CloseNotifier, ok bool) {
	closeNotifier, ok = urw.w.(http.CloseNotifier)
	return
}

func (urw *UtilResponseWriter) CloseNotify() <-chan bool {
	if closeNotifier, ok := urw.w.(http.CloseNotifier); ok {
		return closeNotifier.CloseNotify()
	}
	return nil
}
