package middleware

import (
	"net/http"
	"strings"
)

// CORS returns middleware that sets cross-origin headers only for allowed
// origins, echoing the request Origin so credentialed requests (the
// Authorization header) work. Allowed origins are exact matches or, when suffix
// is non-empty, any origin whose host equals or is a subdomain of the suffix
// (e.g. suffix "fejd.com" matches "https://dragicevic.fejd.com"). Requests from
// disallowed origins pass through without CORS headers, so the browser blocks
// them but non-browser clients are unaffected.
func CORS(allowedOrigins []string, suffix string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[strings.ToLower(strings.TrimRight(o, "/"))] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && originAllowed(origin, allowed, suffix) {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", origin)
				h.Add("Vary", "Origin")
				h.Set("Access-Control-Allow-Credentials", "true")
				h.Set("Access-Control-Expose-Headers", "Link")
				if r.Method == http.MethodOptions {
					h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					h.Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
					h.Set("Access-Control-Max-Age", "300")
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func originAllowed(origin string, allowed map[string]struct{}, suffix string) bool {
	o := strings.ToLower(strings.TrimRight(origin, "/"))
	if _, ok := allowed[o]; ok {
		return true
	}
	if suffix == "" {
		return false
	}
	return hostMatchesSuffix(o, suffix)
}

// hostMatchesSuffix reports whether the host in an Origin header equals the
// suffix or is a subdomain of it. The origin is scheme://host[:port]; the port
// is dropped and the comparison is case-insensitive.
func hostMatchesSuffix(origin, suffix string) bool {
	s := origin
	if i := strings.Index(s, "://"); i >= 0 {
		s = s[i+3:]
	}
	if i := strings.IndexByte(s, '/'); i >= 0 {
		s = s[:i]
	}
	if i := strings.LastIndexByte(s, ':'); i >= 0 {
		s = s[:i]
	}
	host := strings.TrimSuffix(s, ".")
	base := strings.Trim(strings.TrimSuffix(strings.ToLower(suffix), "."), ".")
	if base == "" {
		return false
	}
	return host == base || strings.HasSuffix(host, "."+base)
}
