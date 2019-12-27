package security

import "net/http"

// CorsMiddleware enables CORS Request on server (for development purposes)
func CorsMiddleware(next http.Handler, frameSource *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", *frameSource)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PROPFIND, MKCOL, MOVE, COPY")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-XSRF-TOKEN, Authorization, Depth, Destination")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if req.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, req)
	})
}

// WebSecurityMiddleware adds good practices security headers on http responses
func WebSecurityMiddleware(next http.Handler, frameSource *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000")
		w.Header().Set("Content-Security-Policy", "default-src 'self' blob:; connect-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com https://raw.githubusercontent.com; style-src * 'unsafe-inline'; script-src 'self'; font-src *; frame-src "+*frameSource+"; frame-ancestors "+*frameSource)
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, req)
	})
}
