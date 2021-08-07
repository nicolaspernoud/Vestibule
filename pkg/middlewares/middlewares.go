package middlewares

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/nicolaspernoud/vestibule/pkg/glob"
)

// Cors enables CORS Request on server, CORS are accepted from the allowed domain, its parent domain, and all subdomains
func Cors(next http.Handler, allowedDomain string, port int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		origin := req.Header.Get("Origin")
		if origin != "" {
			if GetFullHostname(allowedDomain, port) == origin || parentDomain(allowedDomain, port) == origin || glob.Glob(subDomainsGlob(allowedDomain, port), origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PROPFIND, MKCOL, MOVE, COPY")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, XSRF-TOKEN, Authorization, Depth, Destination, X-OC-Mtime")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}
		next.ServeHTTP(w, req)
	})
}

func parentDomain(hostname string, port int) string {
	lead := strings.Split(hostname, ".")[0]
	parentHostname := strings.TrimPrefix(hostname, lead+".")
	if port == 80 || port == 443 {
		return "https://" + parentHostname
	}
	return "https://" + parentHostname + ":" + strconv.Itoa(port)
}

func subDomainsGlob(hostname string, port int) string {
	if port == 80 || port == 443 {
		return "https://*." + hostname
	}
	return "https://*." + hostname + ":" + strconv.Itoa(port)
}

/*
**
Indirection layer to manage headers for WebSecurity middleware
**
*/

type webSecurityWriter struct {
	w                     http.ResponseWriter
	source                string
	allowEvalInlineScript bool
	wroteHeader           bool
}

func (s webSecurityWriter) WriteHeader(code int) {
	if !s.wroteHeader {
		s.w.Header().Set("Strict-Transport-Security", "max-age=63072000")
		var inline string
		if s.allowEvalInlineScript {
			inline = "'unsafe-inline' 'unsafe-eval'"
		}
		// Get existing CSP Header
		cspHeader := s.w.Header().Get("Content-Security-Policy")
		if cspHeader != "" { // If it exists, alter it to inject the vestibule main hostname in authorized frame ancestors
			if strings.Contains(cspHeader, "frame-ancestors") {
				cspHeader = strings.Replace(cspHeader, "frame-ancestors", fmt.Sprintf("frame-ancestors %v", s.source), 1)
			} else {
				cspHeader = cspHeader + fmt.Sprintf("; frame-ancestors %v", s.source)
			}
		} else { // If not, forge a default CSP Header
			cspHeader = fmt.Sprintf("default-src %[1]v 'self'; img-src %[1]v 'self' blob: data: ; script-src 'self' %[1]v %[2]v; style-src 'self' 'unsafe-inline'; frame-src %[1]v; frame-ancestors %[1]v", s.source, inline)
		}
		// Set the resulting CSP Header
		s.w.Header().Set("Content-Security-Policy", cspHeader)
		// s.w.Header().Set("X-Frame-Options", "SAMEORIGIN") // Deactivated as browsers take into account that header instead of frame ancestors
		s.w.Header().Set("X-XSS-Protection", "1; mode=block")
		s.w.Header().Set("Referrer-Policy", "strict-origin")
		s.w.Header().Set("X-Content-Type-Options", "nosniff")
		//lint:ignore SA4005 we need to assign true so that when the WriteHeader method will be used again, we won't rewrite security headers
		s.wroteHeader = true
	}
	s.w.WriteHeader(code)
}

func (s webSecurityWriter) Write(b []byte) (int, error) {
	return s.w.Write(b)
}

func (s webSecurityWriter) Header() http.Header {
	return s.w.Header()
}

func (s webSecurityWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := s.w.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("response writer is not an hijacker")
	}
	return hj.Hijack()
}

// WebSecurity adds good practices security headers on http responses
func WebSecurity(next http.Handler, source string, allowEvalInlineScript bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		sw := webSecurityWriter{
			w:                     w,
			source:                source,
			allowEvalInlineScript: allowEvalInlineScript,
			wroteHeader:           false,
		}
		next.ServeHTTP(sw, req)
	})
}

// NoCache disable caching
func NoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", "no-store, must-revalidate")
		next.ServeHTTP(w, req)
	})
}

// GetFullHostname returns the full hostname of the server
func GetFullHostname(hostname string, port int) string {
	if port == 80 || port == 443 {
		return "https://" + hostname
	}
	return "https://" + hostname + ":" + strconv.Itoa(port)
}
