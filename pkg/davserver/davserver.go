package davserver

import (
	"archive/zip"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nicolaspernoud/vestibule/pkg/auth"
	"github.com/nicolaspernoud/vestibule/pkg/common"
	"github.com/nicolaspernoud/vestibule/pkg/log"
	"golang.org/x/net/webdav"
)

// authzFunc creates a middleware to allow access according to a role array
type authzFunc func(http.Handler, []string, bool) http.Handler

// Server implements an http.Handler that acts as an augmented webdav server
type Server struct {
	Mu    sync.RWMutex // guards the fields below
	last  time.Time
	Davs  []*dav
	file  string
	authz authzFunc
}

// NewServer constructs a Server that reads davs from file
func NewServer(file string, authzF authzFunc) (*Server, error) {
	s := new(Server)
	s.authz = authzF
	s.file = file
	if err := s.LoadDavs(); err != nil {
		return nil, err
	}
	return s, nil
}

// ServeHTTP matches the Request with a dav and, if found, serves the request with the dav's handler.
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
	for _, dav := range s.Davs {
		if host == dav.Host {
			return dav.handler
		}
	}
	return nil
}

// LoadDavs tests whether file has been modified since its last invocation and, if so, loads the dav set from file.
func (s *Server) LoadDavs() error {
	fi, err := os.Stat(s.file)
	if err != nil {
		return err
	}
	mtime := fi.ModTime()
	if !mtime.After(s.last) && s.Davs != nil {
		return nil // no change
	}
	davs, err := parseDavs(s.file, s.authz)
	if err != nil {
		return err
	}
	s.Mu.Lock()
	s.last = mtime
	s.Davs = davs
	s.Mu.Unlock()
	return nil
}

// parseDavs reads dav definitions from file, constructs the dav handlers,and returns the resultant davs.
func parseDavs(file string, authz authzFunc) ([]*dav, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var davs []*dav
	if err := json.NewDecoder(f).Decode(&davs); err != nil {
		return nil, err
	}
	for _, r := range davs {
		r.handler = makeHandler(r, authz)
		if r.handler == nil {
			log.Logger.Printf("bad dav: %#v", r)
		}
	}
	return davs, nil
}

// makeHandler constructs the appropriate Handler for the given dav.
func makeHandler(dav *dav, authz authzFunc) http.Handler {
	handler := NewWebDavAug("", dav.Root, dav.Writable)
	if !dav.Secured {
		return handler
	}
	return authz(handler, dav.Roles, true)
}

// WebdavAug represents an augmented webdav which enable download of directories as streamed zip files
type WebdavAug struct {
	prefix     string
	directory  string
	methodMux  map[string]http.Handler
	zipHandler http.Handler
}

// NewWebDavAug create an initialized WebdavAug instance
func NewWebDavAug(prefix string, directory string, canWrite bool) WebdavAug {
	zipH := http.StripPrefix(prefix, &zipHandler{directory})
	davH := &webdav.Handler{
		Prefix:     prefix,
		FileSystem: webdav.Dir(directory),
		LockSystem: webdav.NewMemLS(),
		Logger:     webdavLogger,
	}
	var mMux map[string]http.Handler

	if canWrite {
		mMux = map[string]http.Handler{
			"GET":       davH,
			"OPTIONS":   davH,
			"PROPFIND":  davH,
			"PROPPATCH": davH,
			"MKCOL":     davH,
			"COPY":      davH,
			"MOVE":      davH,
			"LOCK":      davH,
			"UNLOCK":    davH,
			"DELETE":    davH,
			"PUT":       davH,
		}
	} else {
		mMux = map[string]http.Handler{
			"GET":      davH,
			"OPTIONS":  davH,
			"PROPFIND": davH,
		}
	}

	return WebdavAug{
		prefix:     prefix,
		directory:  directory,
		methodMux:  mMux,
		zipHandler: zipH,
	}

}

func (wdaug WebdavAug) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := wdaug.methodMux[r.Method]; ok {
		if r.Method == "GET" {
			// Work out if trying to serve a directory
			ressource := strings.TrimPrefix(r.URL.Path, wdaug.prefix)
			fullName := filepath.Join(wdaug.directory, filepath.FromSlash(path.Clean("/"+ressource)))
			info, err := os.Stat(fullName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !info.IsDir() { // The file will be handled by webdav server
				filename := url.PathEscape(filepath.Base(r.URL.Path))
				_, inline := r.URL.Query()["inline"]
				if !inline {
					w.Header().Set("Content-Disposition", "attachment; filename*="+filename)
				}
			} else {
				h = wdaug.zipHandler
			}
		}
		h.ServeHTTP(w, r)
	} else {
		http.Error(w, "method not allowed : dav is read only", http.StatusMethodNotAllowed)
	}
}

func webdavLogger(r *http.Request, err error) {
	user, err := auth.GetTokenData(r)
	if err != nil && !common.Contains([]string{"PROPFIND", "OPTIONS", "LOCK", "UNLOCK", "GET"}, r.Method) || strings.Contains(user.Login, "_share_") {
		if err != nil {
			log.Logger.Printf("| %v | Webdav access error : [%s] %s, %s | %v | %v", user.Login, r.Method, r.URL, err, r.RemoteAddr, log.GetCityAndCountryFromRequest(r))
		} else {
			log.Logger.Printf("| %v | Webdav access : [%s] %s | %v | %v", user.Login, r.Method, r.URL.Path, r.RemoteAddr, log.GetCityAndCountryFromRequest(r))
		}
	}
}

type zipHandler struct {
	root string
}

func (zh *zipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	webdavLogger(r, nil)
	zipAndServe(w, zh.root, r.URL.Path)
}

func zipAndServe(w http.ResponseWriter, root string, name string) {

	source := filepath.Join(root, filepath.FromSlash(path.Clean("/"+name)))

	size, err := maxZipSize(source)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("content-length", strconv.FormatInt(size, 10))

	archive := zip.NewWriter(w)
	defer archive.Close()

	var rootPath string

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {

		// On root call, set filename and rootPath
		if rootPath == "" {
			rootPath = path
			w.Header().Set("Content-Disposition", "attachment; filename*="+url.PathEscape(info.Name())+".zip")
		}

		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name, err = filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}
		header.Method = zip.Deflate

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func maxZipSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size() + 262144 // Allow 256 kB for zip files overhead (headers, etc.)
		}
		return err
	})
	return size, err
}
