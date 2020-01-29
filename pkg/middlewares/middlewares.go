package middlewares

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/nicolaspernoud/vestibule/pkg/tokens"
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

// WebSecurity adds good practices security headers on http responses
func WebSecurity(next http.Handler, source string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000")
		w.Header().Set("Content-Security-Policy", fmt.Sprintf("default-src %[1]v 'self'; img-src %[1]v blob: 'self'; script-src 'self' %[1]v; style-src 'self' 'unsafe-inline'; frame-src %[1]v; frame-ancestors %[1]v", source))
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, req)
	})
}

// NoCache disable caching
func NoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", "no-store, must-revalidate")
		next.ServeHTTP(w, req)
	})
}

// Encrypt enables body encryption
func Encrypt(next http.Handler, key []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		readBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		encryptedBody, err := tokens.Encrypt(readBody, key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newBody := ioutil.NopCloser(bytes.NewBuffer(encryptedBody))
		r.Body = newBody
		next.ServeHTTP(w, r)
	})
}

// Decrypt enables body decryption
func Decrypt(next http.Handler, key []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wbuff := newDecryptWriter()
		next.ServeHTTP(wbuff, r)
		decryptedData, err := tokens.Decrypt(wbuff.body, key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(decryptedData)
		return
	})
}

type decryptWriter struct {
	headers http.Header
	body    []byte
	status  int
}

func newDecryptWriter() *decryptWriter {
	return &decryptWriter{
		headers: make(http.Header),
	}
}

func (r *decryptWriter) Header() http.Header {
	return r.headers
}

func (r *decryptWriter) Write(p []byte) (int, error) {
	r.body = append(r.body, p...)
	return len(p), nil
}

func (r *decryptWriter) WriteHeader(status int) {
	r.status = status
}

// GetFullHostname returns the full hostname of the server
func GetFullHostname(hostname string, port int) string {
	if port == 80 || port == 443 {
		return "https://" + hostname
	}
	return "https://" + hostname + ":" + strconv.Itoa(port)
}
