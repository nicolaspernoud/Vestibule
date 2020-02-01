package middlewares

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/secure-io/sio-go"
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

// Encrypt enables body encryption (to be used with webdav PUT requests)
func Encrypt(next http.Handler, key []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		block, _ := aes.NewCipher(key)
		gcm, _ := cipher.NewGCM(block)
		stream := sio.NewStream(gcm, sio.BufSize)
		nonce := make([]byte, stream.NonceSize())
		encBody := stream.EncryptReader(r.Body, nonce, nil)
		r.Body = ioutil.NopCloser(encBody)
		next.ServeHTTP(w, r)
	})
}

// Decrypt enables body decryption (to be used with webdav GET requests)
func Decrypt(next http.Handler, key []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decryptWriter := newDecryptWriter(w, key)
		// Force write header to disable content length writing (which will be wrong as it will be the encrypted one)
		w.WriteHeader(200)
		next.ServeHTTP(decryptWriter, r)
		decryptWriter.encWriter.Close()
	})
}

type decryptWriter struct {
	writer    http.ResponseWriter
	encWriter io.WriteCloser
}

func newDecryptWriter(w http.ResponseWriter, key []byte) *decryptWriter {
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	stream := sio.NewStream(gcm, sio.BufSize)
	nonce := make([]byte, stream.NonceSize())
	encWriter := stream.DecryptWriter(w, nonce, nil)
	return &decryptWriter{
		writer:    w,
		encWriter: encWriter,
	}
}

func (r *decryptWriter) Header() http.Header {
	return r.writer.Header()
}

func (r *decryptWriter) Write(p []byte) (int, error) {
	l, err := r.encWriter.Write(p)
	if err != nil {
		panic(err)
	}
	return l, err
}

func (r *decryptWriter) WriteHeader(status int) {
	r.writer.WriteHeader(status)
}

// GetFullHostname returns the full hostname of the server
func GetFullHostname(hostname string, port int) string {
	if port == 80 || port == 443 {
		return "https://" + hostname
	}
	return "https://" + hostname + ":" + strconv.Itoa(port)
}
