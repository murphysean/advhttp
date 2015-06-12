package advhttp

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	CorsOrigin                        = "Origin"
	CorsAccessControlRequestMethod    = "Access-Control-Request-Method"
	CorsAccessControlRequestHeader    = "Access-Control-Request-Header"
	CorsAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	CorsAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	CorsAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	CorsAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	CorsAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	CorsAccessControlMaxAge           = "Access-Control-Max-Age"
)

var (
	CorsDefaultAllowOrigin      = "*"
	CorsDefaultAllowHeaders     = []string{"Location", "Content-Type", "ETag", "Accept-Patch"}
	CorsDefaultAllowMethods     = []string{"OPTIONS", "HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"}
	CorsDefaultExposeHeaders    = []string{"Location", "Content-Type", "ETag", "Accept-Patch"}
	CorsDefaultMaxAge           = int64(1728000)
	CorsDefaultAllowCredentials = true
)

type Cors struct {
	AllowOrigin      string
	AllowHeaders     []string
	AllowMethods     []string
	ExposeHeaders    []string
	MaxAge           int64
	AllowCredentials bool
}

//This function will write out cross origin headers so that javascript clients can call apis.
func (cors *Cors) ProcessCors(w http.ResponseWriter, r *http.Request) {
	//Following this flowchart: http://www.html5rocks.com/static/images/cors_server_flowchart.png
	//Does the request have an Origin Header
	if r.Header.Get(CorsOrigin) == "" {
		//Not a valid CORS request
		return
	}

	//Is the HTTP method an OPTIONS request and does it have a valid Access-Control-Request-Method header?
	if r.Method == "OPTIONS" && r.Header.Get(CorsAccessControlRequestMethod) != "" {
		//Does the request have an Access-Control-Request-Header header?
		if r.Header.Get(CorsAccessControlRequestHeader) != "" {
			//Is the Access-Control-Request-Header header valid? Yes...
			w.Header().Set(CorsAccessControlAllowHeaders, r.Header.Get(CorsAccessControlRequestHeader))
		} else {
			//Set the Access-Control-Allow-Headers response header
			w.Header().Set(CorsAccessControlAllowHeaders, strings.Join(cors.AllowHeaders, ","))
		}
		//Set the Access-Control-Allow-Methods header
		w.Header().Set(CorsAccessControlAllowMethods, strings.Join(cors.AllowMethods, ","))

		//Optional Set the Access-Control-Max-Age response header
		w.Header().Set(CorsAccessControlMaxAge, fmt.Sprintf("%d", cors.MaxAge))
	} else {
		//Actual Request
		w.Header().Set(CorsAccessControlExposeHeaders, strings.Join(cors.ExposeHeaders, ","))
	}

	//Set the Access-Control-Allow-Origin header
	if cors.AllowOrigin == "" {
		w.Header().Set(CorsAccessControlAllowOrigin, r.Header.Get("Origin"))
	} else {
		w.Header().Set(CorsAccessControlAllowOrigin, cors.AllowOrigin)
	}
	//Are cookies allowed?
	w.Header().Set(CorsAccessControlAllowCredentials, fmt.Sprintf("%t", cors.AllowCredentials))
}

var DefaultCors *Cors

func ProcessCors(w http.ResponseWriter, r *http.Request) {
	DefaultCors.ProcessCors(w, r)
}
