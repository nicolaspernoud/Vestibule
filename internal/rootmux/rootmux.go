package rootmux

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/nicolaspernoud/vestibule/pkg/appserver"
	"github.com/nicolaspernoud/vestibule/pkg/auth"
	"github.com/nicolaspernoud/vestibule/pkg/davserver"
	"github.com/nicolaspernoud/vestibule/pkg/middlewares"
	"github.com/nicolaspernoud/vestibule/pkg/onlyoffice"
	"golang.org/x/crypto/acme/autocert"

	"github.com/nicolaspernoud/vestibule/pkg/common"
	"github.com/nicolaspernoud/vestibule/pkg/log"
)

// RootMux represents the main controller of the application
type RootMux struct {
	Mux     http.Handler
	Policy  autocert.HostPolicy
	Manager *auth.Manager
}

// CreateRootMux creates a RootMux
func CreateRootMux(port int, appsFile string, davsFile string, staticDir string) RootMux {
	hostname := os.Getenv("HOSTNAME")
	fullHostname := middlewares.GetFullHostname(hostname, port)
	// Create the app handler
	appServer, err := appserver.NewServer(appsFile, port, hostname, auth.ValidateAuthMiddleware)
	if err != nil {
		log.Logger.Fatal(err)
	}
	// Create the dav handler
	davServer, err := davserver.NewServer(davsFile, auth.ValidateAuthMiddleware)
	if err != nil {
		log.Logger.Fatal(err)
	}
	// Put the two handler together
	adH := &appDavHandler{as: appServer, ds: davServer, dsCORSAllowOrigin: fullHostname, cspSrc: fmt.Sprintf("*.%[1]v:* %[1]v:*", hostname)}
	policy := CreateHostPolicy(hostname, adH)
	// Create the main handler
	mainMux := http.NewServeMux()
	// ALL USERS API ENDPOINTS
	// Authentication endpoints
	m := auth.NewManager()
	mainMux.HandleFunc("/OAuth2Login", m.HandleOAuth2Login)
	mainMux.Handle("/OAuth2Callback", m.HandleOAuth2Callback())
	mainMux.HandleFunc("/Logout", m.HandleLogout)
	mainMux.HandleFunc("/Login", m.HandleInMemoryLogin)
	mainMux.HandleFunc("/onlyoffice", onlyoffice.HandleOpen(fullHostname))
	mainMux.HandleFunc("/onlyoffice/save", onlyoffice.HandleSaveCallback)
	commonMux := http.NewServeMux()
	commonMux.HandleFunc("/apps", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			appServer.ProcessApps(w, req)
			return
		}
		http.Error(w, "method not allowed", 405)
	})
	commonMux.HandleFunc("/davs", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			davServer.ProcessDavs(w, req)
			return
		}
		http.Error(w, "method not allowed", 405)
	})
	mainMux.Handle("/api/common/WhoAmI", auth.ValidateAuthMiddleware(auth.WhoAmI(), []string{"*"}, false))
	commonMux.HandleFunc("/Share", auth.GetShareToken)
	mainMux.Handle("/api/common/", http.StripPrefix("/api/common", auth.ValidateAuthMiddleware(commonMux, []string{"*"}, true)))
	// ADMIN API ENDPOINTS
	adminMux := http.NewServeMux()
	adminMux.Handle("/reload", reload(adH))
	adminMux.HandleFunc("/apps/", appServer.ProcessApps)
	adminMux.HandleFunc("/davs/", davServer.ProcessDavs)
	adminMux.HandleFunc("/users/", auth.ProcessUsers)
	mainMux.Handle("/api/admin/", http.StripPrefix("/api/admin", auth.ValidateAuthMiddleware(adminMux, []string{os.Getenv("ADMIN_ROLE")}, true)))
	// Serve static files falling back to serving index.html
	mainMux.Handle("/", middlewares.NoCache(http.FileServer(&common.FallBackWrapper{Assets: http.Dir(staticDir)})))
	// Put it together into the main handler
	mux := http.NewServeMux()
	mux.Handle(hostname+"/", middlewares.WebSecurity(mainMux, "*."+hostname+":*", false))
	mux.Handle("/", adH)
	return RootMux{mux, policy, &m}
}

type appDavHandler struct {
	as                *appserver.Server
	ds                *davserver.Server
	dsCORSAllowOrigin string
	cspSrc            string
}

func (h *appDavHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	// Some clients include a port in the request host; strip it.
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	for _, a := range h.as.Apps {
		if (host == a.Host) || (strings.Contains(a.Host, "*") && strings.HasSuffix(host, strings.TrimPrefix(a.Host, "*."))) {
			h.as.ServeHTTP(w, r)
			return
		}
	}
	for _, d := range h.ds.Davs {
		if host == d.Host {
			middlewares.Cors(middlewares.WebSecurity(h.ds, h.cspSrc, false), h.dsCORSAllowOrigin).ServeHTTP(w, r)
			return
		}
	}

}

func reload(adh *appDavHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := adh.as.LoadApps()
		if err != nil {
			http.Error(w, err.Error(), 400)
		}
		err = adh.ds.LoadDavs()
		if err != nil {
			http.Error(w, err.Error(), 400)
		} else {
			fmt.Fprintf(w, "apps and davs services reloaded")
		}
	})
}

// CreateHostPolicy implements autocert.HostPolicy
func CreateHostPolicy(hostname string, h *appDavHandler) func(ctx context.Context, host string) error {
	return func(ctx context.Context, host string) error {
		// Appserver
		h.as.Mu.RLock()
		defer h.as.Mu.RUnlock()
		// Check if host is main host
		if host == hostname {
			return nil
		}
		// If not check if the host is in allowed apps
		for _, app := range h.as.Apps {
			if (host == app.Host) || (strings.Contains(app.Host, "*") && strings.HasSuffix(host, strings.TrimPrefix(app.Host, "*."))) {
				return nil
			}
		}
		// Davserver
		h.ds.Mu.RLock()
		defer h.ds.Mu.RUnlock()
		// Check if host is main host
		if host == hostname {
			return nil
		}
		// If not check if the host is in allowed apps
		for _, app := range h.ds.Davs {
			if host == app.Host {
				return nil
			}
		}
		return fmt.Errorf("unrecognized host %q", host)
	}
}
