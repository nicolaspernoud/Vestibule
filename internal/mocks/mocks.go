// Package mocks provide mocks for development purposes (debug mode)
package mocks

import (
	"fmt"
	"net/http"
	"os"

	"github.com/nicolaspernoud/vestibule/pkg/middlewares"
)

var (
	hostname = os.Getenv("HOSTNAME")
	port     int
)

// Init initialize the configuration
func Init(portFromMain int) {
	port = portFromMain
}

// CreateMockOAuth2 creates a mock OAuth2 serve mux for development purposes
func CreateMockOAuth2() *http.ServeMux {
	mux := http.NewServeMux()
	// Returns authorization code back to the user
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		redir := query.Get("redirect_uri") + "?state=" + query.Get("state") + "&code=mock_code"
		http.Redirect(w, r, redir, http.StatusFound)
	})
	// Returns authorization code back to the user, but without the provided state
	mux.HandleFunc("/auth-wrong-state", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		redir := query.Get("redirect_uri") + "?state=" + "a-random-state" + "&code=mock_code"
		http.Redirect(w, r, redir, http.StatusFound)
	})

	// Returns access token back to the user
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.Write([]byte("access_token=mocktoken&scope=user&token_type=bearer"))
	})
	// Returns userinfo back to the user
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"displayName": "Us ER",
			"memberOf": [
				"CN=USERS",
				"CN=OTHER_GROUP"
			],
			"id": "1000",
			"login": "USER"
		}`))
	})
	// Returns userinfo back to the user (with an admin user)
	mux.HandleFunc("/admininfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"displayName": "Ad MIN",
			"memberOf": [
				"CN=ADMINS",
				"CN=OTHER_GROUP"
			],
			"id": "1",
			"login": "ADMIN"
		}`))
	})
	// Logout
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Logout OK")
	})

	return mux
}

// CreateMockAPI creates a mock OAuth2 serve mux for development purposes
func CreateMockAPI() *http.ServeMux {
	mux := http.NewServeMux()
	// Returns authorization code back to the user
	mux.Handle("/", middlewares.Cors(func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors http://www.example.com")
			w.Write([]byte(`{
				"foo": "bar",
				"bar": "foo"
			}`))
		})
	}(), hostname, port))
	return mux
}
