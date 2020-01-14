/*

This package is based upon https://github.com/nf/webfront (Copyright 2011 Google Inc.)

*/

package appserver

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	port        int
	frameSource string
)

// authzFunc creates a middleware to allow access according to a role array
type authzFunc func(http.Handler, []string, bool) http.Handler

// Server implements an http.Handler that acts as either a reverse proxy or a simple file server, as determined by a rule set.
type Server struct {
	Mu    sync.RWMutex // guards the fields below
	last  time.Time
	Apps  []*app
	file  string
	authz authzFunc
}

// NewServer constructs a Server that reads apps from file
func NewServer(file string, portFromMain int, frameSourceFromMain string, authzFromMain authzFunc) (*Server, error) {
	port = portFromMain
	frameSource = frameSourceFromMain
	s := new(Server)
	s.authz = authzFromMain
	s.file = file
	if err := s.LoadApps(); err != nil {
		return nil, err
	}
	return s, nil
}

// ServeHTTP matches the Request with a app and, if found, serves the request with the app's handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h := s.handler(r); h != nil {
		h.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Not found.", http.StatusNotFound)
}

// handler returns the appropriate Handler for the given Request,
// or nil if none found.
func (s *Server) handler(req *http.Request) http.Handler {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	host := req.Host
	// Some clients include a port in the request host; strip it.
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	for _, app := range s.Apps {
		// Standard case
		if !strings.HasPrefix(app.Host, "*.") && host == app.Host {
			return app.handler
		}
		// Wilcard case
		if strings.HasPrefix(app.Host, "*.") && (host == app.Host || host == strings.TrimPrefix(app.Host, "*.") || strings.HasSuffix(host, strings.TrimPrefix(app.Host, "*"))) {
			return app.handler
		}
	}
	return nil
}

// LoadApps tests whether file has been modified since its last invocation and, if so, loads the app set from file.
func (s *Server) LoadApps() error {
	fi, err := os.Stat(s.file)
	if err != nil {
		return err
	}
	mtime := fi.ModTime()
	if !mtime.After(s.last) && s.Apps != nil {
		return nil // no change
	}
	apps, err := parseApps(s.file, s.authz)
	if err != nil {
		return err
	}
	s.Mu.Lock()
	s.last = mtime
	s.Apps = apps
	s.Mu.Unlock()
	return nil
}

// parseApps reads app definitions from file, constructs the app handlers,and returns the resultant apps.
func parseApps(file string, authz authzFunc) ([]*app, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var apps []*app
	if err := json.NewDecoder(f).Decode(&apps); err != nil {
		return nil, err
	}
	for _, r := range apps {
		r.handler = makeHandler(r, authz)
		if r.handler == nil {
			log.Printf("bad app: %#v", r)
		}
	}
	return apps, nil
}

// makeHandler constructs the appropriate Handler for the given app.
func makeHandler(app *app, authz authzFunc) http.Handler {
	var handler http.Handler
	if fwdTo := app.ForwardTo; app.IsProxy && fwdTo != "" {
		fwdFrom := strings.TrimPrefix(app.Host, "*.")
		handler = &httputil.ReverseProxy{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Director: func(req *http.Request) {
				// Set the correct scheme to the request
				if !strings.HasPrefix(fwdTo, "http") {
					req.URL.Scheme = "http"
					req.URL.Host = fwdTo
				} else {
					fwdToSplit := strings.Split(fwdTo, "://")
					req.URL.Scheme = fwdToSplit[0]
					req.URL.Host = fwdToSplit[1]
				}
				// Rewrite host header if the proxy is not to a local service
				if !strings.Contains(fwdTo, ":") {
					req.Host = fwdTo
				}
				if app.Login != "" && app.Password != "" {
					req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(app.Login+":"+app.Password)))
				}
			},
			ModifyResponse: func(res *http.Response) error {
				u, err := res.Location()
				if err == nil {
					// Alter the redirect location if the redirection is relative to the proxied host
					if strings.Contains(u.Host, fwdTo) {
						u.Scheme = "https"
						u.Host = fwdFrom + ":" + strconv.Itoa(port)
					}
					res.Header.Set("Location", u.String())
				}
				res.Header.Set("Content-Security-Policy", "frame-ancestors "+frameSource)
				res.Header.Set("X-Frame-Options", "DENY")
				return nil
			},
		}
	} else if d := app.Serve; !app.IsProxy && d != "" {
		handler = http.FileServer(http.Dir(d))
	}
	if !app.Secured || handler == nil {
		return handler
	}
	return authz(handler, app.Roles, false)
}
