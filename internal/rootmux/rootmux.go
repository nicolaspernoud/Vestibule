package rootmux

import (
	"fmt"
	"net/http"
	"os"

	"github.com/nicolaspernoud/vestibule/pkg/appserver"
	"github.com/nicolaspernoud/vestibule/pkg/auth"
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
func CreateRootMux(port int, appsFile string, staticDir string) RootMux {
	hostname := os.Getenv("HOSTNAME")
	// Create the app handler
	appServer, err := appserver.NewServer(appsFile, port, hostname, hostname, auth.ValidateJWTAndRolesMiddleware)
	if err != nil {
		log.Logger.Fatal(err)
	}
	var appHandler http.Handler = appServer
	appserver.Init(appsFile)
	// Create the main handler
	mainMux := http.NewServeMux()
	// ALL USERS API ENDPOINTS
	// Authentication endpoints
	m := auth.NewManager()
	mainMux.HandleFunc("/OAuth2Login", m.HandleOAuth2Login)
	mainMux.Handle("/OAuth2Callback", m.HandleOAuth2Callback())
	mainMux.HandleFunc("/Logout", m.HandleLogout)
	mainMux.HandleFunc("/Login", m.HandleInMemoryLogin)
	commonMux := http.NewServeMux()
	commonMux.HandleFunc("/apps", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			appserver.ProcessApps(w, req)
			return
		}
		http.Error(w, "method not allowed", 405)
	})
	commonMux.HandleFunc("/WhoAmI", auth.WhoAmI)
	mainMux.Handle("/api/common/", http.StripPrefix("/api/common", auth.ValidateJWTAndRolesMiddleware(commonMux, []string{os.Getenv("COMMON_ROLE")})))
	// ADMIN API ENDPOINTS
	adminMux := http.NewServeMux()
	adminMux.Handle("/apps/reload", reloadApps(appServer, appsFile))
	adminMux.HandleFunc("/apps/", appserver.ProcessApps)
	adminMux.HandleFunc("/users/", auth.ProcessUsers)
	mainMux.Handle("/api/admin/", http.StripPrefix("/api/admin", auth.ValidateJWTAndRolesMiddleware(adminMux, []string{os.Getenv("ADMIN_ROLE")})))
	// Serve static files falling back to serving index.html
	mainMux.Handle("/", http.FileServer(&common.FallBackWrapper{Assets: http.Dir(staticDir)}))
	// Put it together into the main handler
	mux := http.NewServeMux()
	mux.Handle(hostname+"/", mainMux)
	mux.Handle("/", appHandler)
	return RootMux{mux, appServer.HostPolicy, &m}
}

func reloadApps(appServer *appserver.Server, appFile string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := appServer.LoadApps(appFile)
		if err != nil {
			http.Error(w, err.Error(), 400)
		} else {
			fmt.Fprintf(w, "apps reloaded")
		}
	})
}
