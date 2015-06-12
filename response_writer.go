package advhttp

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

type ResponseWriter struct {
	w http.ResponseWriter

	length int64
	status int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w: w, length: 0, status: 200}
}

func (trw *ResponseWriter) Header() http.Header {
	return trw.w.Header()
}

func (trw *ResponseWriter) WriteHeader(status int) {
	trw.status = status
	trw.w.WriteHeader(status)
}

func (trw *ResponseWriter) Write(bytes []byte) (int, error) {
	n, err := trw.w.Write(bytes)
	trw.length += int64(n)
	return n, err
}

func (trw *ResponseWriter) Length() int64 {
	return trw.length
}

func (trw *ResponseWriter) Status() int {
	return trw.status
}

func (trw *ResponseWriter) GetFlusher() (flusher http.Flusher, ok bool) {
	flusher, ok = trw.w.(http.Flusher)
	return
}

func (trw *ResponseWriter) Flush() {
	if flusher, ok := trw.w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (trw *ResponseWriter) GetHijacker() (hijacker http.Hijacker, ok bool) {
	hijacker, ok = trw.w.(http.Hijacker)
	return
}

func (trw *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := trw.w.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.New("Couldn't cast responsewriter to hijacker")
}

func (trw *ResponseWriter) GetCloseNotifier() (closeNotifier http.CloseNotifier, ok bool) {
	closeNotifier, ok = trw.w.(http.CloseNotifier)
	return
}

func (trw *ResponseWriter) CloseNotify() <-chan bool {
	if closeNotifier, ok := trw.w.(http.CloseNotifier); ok {
		return closeNotifier.CloseNotify()
	}
	return nil
}

func (trw *ResponseWriter) LogCommonExtended(r *http.Request) string {
	return LogCommonExtended(trw, r)
}

func (trw *ResponseWriter) LogCommonExtendedForwarded(r *http.Request) string {
	return LogCommonExtendedForwarded(trw, r)
}
