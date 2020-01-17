package davserver

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"sort"

	"github.com/nicolaspernoud/vestibule/pkg/common"
)

// Dav represents a webdav file service
type Dav struct {
	ID       int      `json:"id"`
	Host     string   `json:"host"`               // to match against request Host header
	Root     string   `json:"root"`               // the file system directory to serve the webdav from
	Writable bool     `json:"writable,omitempty"` // whether if the webdav is writable (default to read only)
	Name     string   `json:"name,omitempty"`     // name of the file service
	Icon     string   `json:"icon,omitempty"`     // icon to display
	Color    string   `json:"color,omitempty"`    // icon's color
	Secured  bool     `json:"secured"`            // true if the handler is secured with auth
	Roles    []string `json:"roles,omitempty"`    // Roles allowed to access the file service
}

type dav struct {
	Dav
	handler http.Handler
}

// ByID implements sort.Interface for []Dav based on the Id field
type ByID []Dav

func (a ByID) Len() int           { return len(a) }
func (a ByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByID) Less(i, j int) bool { return a[i].ID < a[j].ID }

// ProcessDavs processes davs regarding of HTTP method
func (s *Server) ProcessDavs(w http.ResponseWriter, req *http.Request) {
	switch method := req.Method; method {
	case "GET":
		s.SendDavs(w, req)
	case "POST":
		s.AddDav(w, req)
	case "DELETE":
		s.DeleteDav(w, req)
	default:
		http.Error(w, "method not allowed", 400)
	}
}

// SendDavs send davs as response from an http requests
func (s *Server) SendDavs(w http.ResponseWriter, req *http.Request) {
	var davs []Dav
	err := common.Load(s.file, &davs)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	json.NewEncoder(w).Encode(davs)
}

// AddDav adds an dav
func (s *Server) AddDav(w http.ResponseWriter, req *http.Request) {
	var davs []Dav
	err := common.Load(s.file, &davs)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if req.Body == nil {
		http.Error(w, "please send a request body", 400)
		return
	}
	var newDav Dav
	err = json.NewDecoder(req.Body).Decode(&newDav)
	if _, ok := err.(*json.UnmarshalTypeError); !ok && err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	// Add the dav only if the id doesn't exists yet
	isNew := true
	for idx, val := range davs {
		if val.ID == newDav.ID {
			davs[idx] = newDav
			isNew = false
			break
		}
	}
	if isNew {
		davs = append(davs, newDav)
		sort.Sort(ByID(davs))
	}
	err = common.Save(s.file, &davs)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.SendDavs(w, req)
}

// DeleteDav adds an dav
func (s *Server) DeleteDav(w http.ResponseWriter, req *http.Request) {
	var davs []Dav
	err := common.Load(s.file, &davs)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	pathElements := strings.Split(req.URL.Path, "/")
	idx, err := strconv.Atoi(pathElements[len(pathElements)-1])
	if err != nil {
		http.Error(w, "please provide an dav index", 400)
		return
	}
	// Add the dav only if the name doesn't exists yet
	newDavs := davs[:0]
	for _, dav := range davs {
		if dav.ID != idx {
			newDavs = append(newDavs, dav)
		}
	}
	err = common.Save(s.file, &newDavs)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.SendDavs(w, req)
}