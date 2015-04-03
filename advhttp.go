package advhttp

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
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

func LogApache(trw *ResponseWriter, r *http.Request) string {
	remoteAddr := r.RemoteAddr
	if remoteAddr == "" {
		remoteAddr = "-"
	}
	method := r.Method
	userId := r.Header.Get("User-Id")
	if userId == "" {
		userId = "-"
	}
	referer := r.Referer()
	if referer == "" {
		referer = "-"
	}
	userAgent := r.UserAgent()
	if userAgent == "" {
		userAgent = "-"
	}
	return fmt.Sprintf("%v - %v [%v] \"%v %v %v\" %v %v %v %v\n", remoteAddr, userId, time.Now().UTC().Format(http.TimeFormat), method, r.URL.String(), r.Proto, trw.status, trw.length, referer, userAgent)
}

func LogApacheWithHeader(trw *ResponseWriter, r *http.Request, url, header string) string {
	remoteAddr := r.RemoteAddr
	if remoteAddr == "" {
		remoteAddr = "-"
	}
	method := r.Method
	userId := r.Header.Get("User-Id")
	if userId == "" {
		userId = "-"
	}
	referer := r.Referer()
	if referer == "" {
		referer = "-"
	}
	userAgent := r.UserAgent()
	if userAgent == "" {
		userAgent = "-"
	}
	extraHeader := r.Header.Get(header)
	if r.Header.Get(header) == "" {
		extraHeader = "-"
	}

	return fmt.Sprintf("%v - %v [%v] \"%v %v %v\" %v %v %v %v %v", remoteAddr, userId, time.Now().UTC().Format(http.TimeFormat), method, url, r.Proto, trw.status, trw.length, referer, userAgent, extraHeader)
}

//This function will write out cross origin headers so that javascript clients can call apis.
func ProcessCors(w http.ResponseWriter, r *http.Request) {
	//Following this flowchart: http://www.html5rocks.com/static/images/cors_server_flowchart.png
	//Does the request have an Origin Header
	if r.Header.Get("Origin") == "" {
		//Not a valid CORS request
		return
	}

	//Is the HTTP method an OPTIONS request?
	if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
		//Is the Access-Control-Request-Method header valid? Yes...
		//Does the request have an Access-Control-Request-Header header?
		if r.Header.Get("Access-Control-Request-Header") != "" {
			//Is the Access-Control-Request-Header header valid? Yes...
			w.Header().Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Header"))
		} else {
			//Set the Access-Control-Allow-Headers response header
			w.Header().Set("Access-Control-Allow-Headers", "Location, Content-Type, ETag, Accept-Patch")
		}
		//Set the Access-Control-Allow-Methods header
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, HEAD, GET, POST, PUT, PATCH, DELETE")

		//Optional Set the Access-Control-Max-Age response header
		w.Header().Set("Access-Control-Max-Age", "1728000")
	} else {
		//Actual Request
		w.Header().Set("Access-Control-Expose-Headers", "Location, Content-Type, ETag, Accept-Patch")
	}

	//Set the Access-Control-Allow-Origin header
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	//Are cookies allowed?
	//w.Header().Set("Access-Control-Allow-Credentials", "true")
}

//BearerAuth is a function that will pull an access token out of the Authorization header
//it will return the bearer token if found, and ok will tell you whether it was able to find
//the token or not. This function will look for the token in the query params, as well as
//the headers.
func BearerAuth(r *http.Request) (bearerToken string, ok bool) {
	ok = false
	if values, o := r.URL.Query()["access_token"]; o && len(values) > 0 {
		bearerToken = values[0]
		ok = true
	} else if authorization := r.Header.Get("Authorization"); strings.HasPrefix(authorization, "Bearer ") {
		bearerToken = r.Header.Get("Authorization")[7:]
		ok = true
	}
	return
}

//This function will take in the accept header string from a inbound request and determine if
//application/json is an acceptable response for the request. It ignores any priorities that
//the requester has, and if they don't include an accept header it will treat it as if they
//had just used Accept: */*
func IsJSONAnAcceptableResponse(acceptHeader string) bool {
	//No accept header (or empty) means */*
	if acceptHeader == "" {
		return true
	}

	//Since the browser can use something like this: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
	//Acceptable content types are application/json application/* and */*
	for _, ent := range strings.Split(acceptHeader, ",") {
		parts := strings.Split(ent, ";")
		if len(parts) >= 1 {
			if parts[0] == "application/json" || parts[0] == "application/*" || parts[0] == "*/*" {
				return true
			}
		}
	}
	return false
}
