package advhttp

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

func init() {
	DefaultCors = new(Cors)
	DefaultCors.AllowOrigin = ""
	DefaultCors.AllowHeaders = CorsDefaultAllowHeaders
	DefaultCors.AllowMethods = CorsDefaultAllowMethods
	DefaultCors.ExposeHeaders = CorsDefaultExposeHeaders
	DefaultCors.MaxAge = CorsDefaultMaxAge
	DefaultCors.AllowCredentials = CorsDefaultAllowCredentials

	DefaultHsts = new(Hsts)
	DefaultHsts.MaxAge = HstsDefaultMaxAge
	DefaultHsts.IncludeSubDomains = HstsDefaultIncludeSubDomains
	DefaultHsts.Preload = HstsDefaultPreload
}

func LogApache(trw *ResponseWriter, r *http.Request) string {
	return logWithOptions(trw, r, false, 0)
}

func LogCommonExtended(trw *ResponseWriter, r *http.Request) string {
	return logWithOptions(trw, r, false, 0)
}

func LogCommonExtendedForwarded(trw *ResponseWriter, r *http.Request) string {
	return logWithOptions(trw, r, true, 0)
}

func LogWithOptions(trw *ResponseWriter, r *http.Request, useXForwarded bool, duration time.Duration) string {
	return logWithOptions(trw, r, useXForwarded, duration)
}

func logWithOptions(trw *ResponseWriter, r *http.Request, useXForwarded bool, duration time.Duration) string {
	remoteAddr := r.RemoteAddr
	if r.Header.Get("X-Forwarded-For") != "" && useXForwarded {
		if fwds := strings.Split(r.Header.Get("X-Forwarded-For"), ","); len(fwds) > 0 {
			remoteAddr = strings.TrimSpace(fwds[0])
		}
	}
	if remoteHost, _, err := net.SplitHostPort(remoteAddr); err == nil {
		remoteAddr = remoteHost
	}
	if remoteAddr == "" {
		remoteAddr = "-"
	}
	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	if r.Header.Get("X-Forwarded-Proto") != "" && useXForwarded {
		proto = r.Header.Get("X-Forwarded-Proto")
	}
	host := r.Host
	if r.Header.Get("X-Forwarded-Host") != "" && useXForwarded {
		host = r.Header.Get("X-Forwarded-Host")
	}
	if host == "" {
		host = "-"
	}
	method := r.Method
	userId := r.Header.Get("X-User-Id")
	if userId == "" {
		userId = "-"
	}
	clientId := r.Header.Get("X-Client-Id")
	if clientId == "" {
		clientId = "-"
	}
	dur := "-"
	if duration != 0 {
		dur = duration.String()
	}
	referer := r.Referer()
	if referer == "" {
		referer = "-"
	}
	userAgent := r.UserAgent()
	if userAgent == "" {
		userAgent = "-"
	}
	return fmt.Sprintf("%v %v %v [%v] %v %v \"%v %v %v\" %v %v %v \"%v\" \"%v\"\n", remoteAddr, clientId, userId, time.Now().UTC().Format(time.RFC3339Nano), proto, host, method, r.URL.String(), r.Proto, trw.status, trw.length, dur, referer, userAgent)
}

// BearerAuth is a function that will pull an access token out of the Authorization header
// it will return the bearer token if found, and ok will tell you whether it was able to find
// the token or not. This function will look for the token in the query params, as well as
// the headers.
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

// This function will take in the accept header string from a inbound request and determine if
// application/json is an acceptable response for the request. It ignores any priorities that
// the requester has, and if they don't include an accept header it will treat it as if they
// had just used Accept: */*
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
