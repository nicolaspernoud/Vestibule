package middlewares

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// Cors enables CORS Request on server (for development purposes)
func Cors(next http.Handler, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PROPFIND, MKCOL, MOVE, COPY")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, XSRF-TOKEN, Authorization, Depth, Destination")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		next.ServeHTTP(w, req)
	})
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
	if s.wroteHeader == false {
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
			cspHeader = fmt.Sprintf("default-src %[1]v 'self'; img-src %[1]v 'self' blob: data: ; script-src 'self' %[1]v %[2]v; style-src 'self' 'unsafe-inline'; frame-src http: %[1]v; frame-ancestors %[1]v", s.source, inline)
		}
		// Set the resulting CSP Header
		s.w.Header().Set("Content-Security-Policy", cspHeader)
		s.w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		s.w.Header().Set("X-XSS-Protection", "1; mode=block")
		s.w.Header().Set("Referrer-Policy", "strict-origin")
		s.w.Header().Set("X-Content-Type-Options", "nosniff")
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
