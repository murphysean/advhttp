package advhttp

import (
	"net"
	"net/http"
	"strings"
)

func AddOutboundHeaders(r *http.Request, host string, via string) {
	r.Header.Set("Host", host)
	r.Host = host
	if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		// If we aren't the first proxy retain prior
		// X-Forwarded-For information as a comma+space
		// separated list and fold multiple headers into one.
		if prior, ok := r.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		r.Header.Set("X-Forwarded-For", clientIP)
	}

	if r.Header.Get("X-Forwarded-Host") == "" {
		r.Header.Set("X-Forwarded-Host", r.Host)
	}

	if r.Header.Get("X-Forwarded-Proto") == "" {
		proto := "http"
		if r.TLS != nil {
			proto = "https"
		}
		r.Header.Set("X-Forwarded-Proto", proto)
	}

	// If we aren't the first proxy retain prior
	// Via information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := r.Header["Via"]; ok {
		via = strings.Join(prior, ", ") + ", " + via
	}
	r.Header.Set("Via", via)
}
