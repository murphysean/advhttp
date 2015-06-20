package advhttp

import (
	"fmt"
	"net/http"
)

const (
	HstsStrictTransportSecurity = "Strict-Transport-Security"
	HstsIncludeSubDomains       = "includeSubDomains"
	HstsPreload                 = "preload"
)

var (
	HstsDefaultMaxAge            = int64(31536000)
	HstsDefaultIncludeSubDomains = false
	HstsDefaultPreload           = false
)

type Hsts struct {
	// Sets the number of seconds that clients will use sts (always https)
	MaxAge int64
	// Enables sts on all subdomains as well
	IncludeSubDomains bool
	// Signals that the site should be added to the browsers preload list
	Preload bool
}

func (hsts *Hsts) ProcessHsts(w http.ResponseWriter, r *http.Request) {
	writeHsts := false
	if r.TLS != nil {
		writeHsts = true
	}
	if r.Header.Get("X-Forwarded-Proto") != "" {
		writeHsts = true
	}
	if writeHsts && hsts.IncludeSubDomains && hsts.Preload {
		w.Header().Set(HstsStrictTransportSecurity, fmt.Sprintf("max-age=%d; %v; %v", hsts.MaxAge, HstsIncludeSubDomains, HstsPreload))
		return
	}
	if writeHsts && hsts.IncludeSubDomains {
		w.Header().Set(HstsStrictTransportSecurity, fmt.Sprintf("max-age=%d; %d", hsts.MaxAge, HstsIncludeSubDomains))
		return
	}
	if writeHsts {
		w.Header().Set(HstsStrictTransportSecurity, fmt.Sprintf("max-age=%d", hsts.MaxAge))
	}
}

var DefaultHsts *Hsts

func ProcessHsts(w http.ResponseWriter, r *http.Request) {
	DefaultHsts.ProcessHsts(w, r)
}
