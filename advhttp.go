package advhttp

import (
	"fmt"
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

func LogApache(trw *ResponseWriter, r *http.Request) string {
	var userId string = ""
	if r.Header.Get("User-Id") == "" {
		userId = "-"
	}
	return fmt.Sprintf("%v - %v [%v] \"%v %v %v\" %v %v %v %v\n", r.RemoteAddr, userId, time.Now().UTC().Format(http.TimeFormat), r.Method, r.URL.String(), r.Proto, trw.status, trw.length, r.Referer(), r.UserAgent())
}

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

func BearerAuth(r *http.Request) (bearerToken string) {
	if values, ok := r.URL.Query()["access_token"]; ok && len(values) > 0 {
		bearerToken = values[0]
	} else if authorization := r.Header.Get("Authorization"); strings.HasPrefix(authorization, "Bearer ") {
		bearerToken = r.Header.Get("Authorization")[7:]
	}
	return
}
