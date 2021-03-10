package appserver

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"sort"

	"github.com/nicolaspernoud/vestibule/pkg/common"
)

// App represents a app serving static content proxying a web server
type App struct {
	ID              int      `json:"id"`
	Name            string   `json:"name,omitempty"`            // name of the app
	Icon            string   `json:"icon,omitempty"`            // icon to display
	Color           string   `json:"color,omitempty"`           // icon's color
	IsProxy         bool     `json:"isProxy"`                   // true if reverse proxy
	Host            string   `json:"host"`                      // to match against request Host header
	ForwardTo       string   `json:"forwardTo,omitempty"`       // non-empty if reverse proxy
	Serve           string   `json:"serve,omitempty"`           // non-empty if static server
	Secured         bool     `json:"secured"`                   // true if the handler is secured with auth
	Login           string   `json:"login,omitempty"`           // Basic auth login for automatic login
	Password        string   `json:"password,omitempty"`        // Basic auth password for automatic login
	OpenPath        string   `json:"openpath,omitempty"`        // Opened path on iframe preview
	Roles           []string `json:"roles,omitempty"`           // Roles allowed to access the app
	SecurityHeaders bool     `json:"securityheaders,omitempty"` // Inject security headers (CSP, STS, etc.)
	CachePattern    []string `json:"cachepattern,omitempty"`    // Cache routes corresponding to that pattern
	CacheDuration   int      `json:"cacheduration,omitempty"`   // Cache routes for seconds
}

type app struct {
	App
	handler http.Handler
}

// ByID implements sort.Interface for []App based on the Id field
type ByID []App

func (a ByID) Len() int           { return len(a) }
func (a ByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByID) Less(i, j int) bool { return a[i].ID < a[j].ID }

// ProcessApps processes apps regarding of HTTP method
func (s *Server) ProcessApps(w http.ResponseWriter, req *http.Request) {
	switch method := req.Method; method {
	case "GET":
		s.SendApps(w, req)
	case "POST":
		s.AddApp(w, req)
	case "DELETE":
		s.DeleteApp(w, req)
	default:
		http.Error(w, "method not allowed", http.StatusBadRequest)
	}
}

// SendApps send apps as response from an http requests
func (s *Server) SendApps(w http.ResponseWriter, req *http.Request) {
	var apps []App
	err := common.Load(s.file, &apps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(apps)
}

// AddApp adds an app
func (s *Server) AddApp(w http.ResponseWriter, req *http.Request) {
	var apps []App
	err := common.Load(s.file, &apps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Body == nil {
		http.Error(w, "please send a request body", http.StatusBadRequest)
		return
	}
	var newApp App
	err = json.NewDecoder(req.Body).Decode(&newApp)
	if _, ok := err.(*json.UnmarshalTypeError); !ok && err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Add the app only if the id doesn't exists yet
	isNew := true
	for idx, val := range apps {
		if val.ID == newApp.ID {
			apps[idx] = newApp
			isNew = false
			break
		}
	}
	if isNew {
		apps = append(apps, newApp)
		sort.Sort(ByID(apps))
	}
	err = common.Save(s.file, &apps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.SendApps(w, req)
}

// DeleteApp adds an app
func (s *Server) DeleteApp(w http.ResponseWriter, req *http.Request) {
	var apps []App
	err := common.Load(s.file, &apps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pathElements := strings.Split(req.URL.Path, "/")
	idx, err := strconv.Atoi(pathElements[len(pathElements)-1])
	if err != nil {
		http.Error(w, "please provide an app index", http.StatusBadRequest)
		return
	}
	// Add the app only if the name doesn't exists yet
	newApps := apps[:0]
	for _, app := range apps {
		if app.ID != idx {
			newApps = append(newApps, app)
		}
	}
	err = common.Save(s.file, &newApps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.SendApps(w, req)
}
