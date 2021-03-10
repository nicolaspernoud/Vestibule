package davserver

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"sort"

	"github.com/nicolaspernoud/vestibule/pkg/auth"
	"github.com/nicolaspernoud/vestibule/pkg/common"
	"github.com/nicolaspernoud/vestibule/pkg/du"
)

// Dav represents a webdav file service
type Dav struct {
	ID                   int      `json:"id"`
	Host                 string   `json:"host"`                 // to match against request Host header
	Root                 string   `json:"root"`                 // the file system directory to serve the webdav from
	Writable             bool     `json:"writable,omitempty"`   // whether if the webdav is writable (default to read only)
	Name                 string   `json:"name,omitempty"`       // name of the file service
	Icon                 string   `json:"icon,omitempty"`       // icon to display
	Color                string   `json:"color,omitempty"`      // icon's color
	Secured              bool     `json:"secured"`              // true if the handler is secured with auth
	Roles                []string `json:"roles,omitempty"`      // Roles allowed to access the file service
	EncryptionPassphrase string   `json:"passphrase,omitempty"` // passphrase to encrypt data
	UsedGB               uint64   `json:"usedgb,omitempty"`
	TotalGB              uint64   `json:"totalgb,omitempty"`
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
		http.Error(w, "method not allowed", http.StatusBadRequest)
	}
}

// SendDavs send davs as response from an http requests
func (s *Server) SendDavs(w http.ResponseWriter, req *http.Request) {
	// Get user from request
	user, err := auth.GetTokenData(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var davs []Dav
	err = common.Load(s.file, &davs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for i, dav := range davs {
		usage, err := du.NewDiskUsage(dav.Root)
		if err == nil {
			dav.UsedGB = usage.Used() / du.GB
			dav.TotalGB = usage.Size() / du.GB
		}
		// Do not leak encryption passphrase to non admins users
		if !user.IsAdmin {
			dav.EncryptionPassphrase = ""
		}
		davs[i] = dav
	}
	json.NewEncoder(w).Encode(davs)
}

// AddDav adds an dav
func (s *Server) AddDav(w http.ResponseWriter, req *http.Request) {
	var davs []Dav
	err := common.Load(s.file, &davs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Body == nil {
		http.Error(w, "please send a request body", http.StatusBadRequest)
		return
	}
	var newDav Dav
	err = json.NewDecoder(req.Body).Decode(&newDav)
	if _, ok := err.(*json.UnmarshalTypeError); !ok && err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.SendDavs(w, req)
}

// DeleteDav adds an dav
func (s *Server) DeleteDav(w http.ResponseWriter, req *http.Request) {
	var davs []Dav
	err := common.Load(s.file, &davs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pathElements := strings.Split(req.URL.Path, "/")
	idx, err := strconv.Atoi(pathElements[len(pathElements)-1])
	if err != nil {
		http.Error(w, "please provide an dav index", http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.SendDavs(w, req)
}
