package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nicolaspernoud/vestibule/pkg/tokens"
)

type key int

const (
	authTokenKey string = "auth_token"
	// ContextData is the user
	ContextData key = 0
)

// User represents a logged in user
type User struct {
	ID           string   `json:"id,omitempty"`
	Login        string   `json:"login"`
	DisplayName  string   `json:"displayName,omitempty"`
	Roles        []string `json:"memberOf"`
	IsAdmin      bool     `json:"isAdmin,omitempty"`
	Name         string   `json:"name,omitempty"`
	Surname      string   `json:"surname,omitempty"`
	PasswordHash string   `json:"passwordHash,omitempty"`
	Password     string   `json:"password,omitempty"`
}

// TokenData represents the data held into a token
type TokenData struct {
	User
	URL              string `json:"url,omitempty"`
	ReadOnly         bool   `json:"readonly,omitempty"`
	SharingUserLogin string `json:"sharinguserlogin,omitempty"`
	XSRFToken        string `json:"xsrftoken,omitempty"`
}

// ValidateAuthMiddleware validates that the token is valid and that the user has the correct roles
func ValidateAuthMiddleware(next http.Handler, allowedRoles []string, checkXSRF bool) http.Handler {
	roleChecker := func(w http.ResponseWriter, r *http.Request) {
		user := TokenData{}
		checkXSRF, err := tokens.Manager.ExtractAndValidateToken(r, authTokenKey, &user, checkXSRF)
		// Handle CORS preflight requests
		if err != nil && r.Method == "OPTIONS" {
			// Handle GIO preflight requests
			if strings.Contains(r.UserAgent(), "vfs") || strings.Contains(r.UserAgent(), "Microsoft-WebDAV") {
				w.Header().Set("WWW-Authenticate", `Basic realm="server"`)
				http.Error(w, "webdav client authentication", 401)
			}
			return
		}
		if err != nil {
			redirectTo := os.Getenv("HOSTNAME")
			_, port, perr := net.SplitHostPort(r.Host)
			if perr == nil {
				redirectTo += ":" + port
			}
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusUnauthorized)
			responseContent := fmt.Sprintf("error extracting token: %v<meta http-equiv=\"Refresh\" content=\"0; url=https://%v/#login\"/>", err.Error(), redirectTo)
			fmt.Fprintf(w, responseContent)
			return
		}
		// Check XSRF Token
		if checkXSRF && r.Header.Get("XSRF-TOKEN") != user.XSRFToken {
			http.Error(w, "XSRF protection triggered", 401)
			return
		}
		err = checkUserHasRole(user, allowedRoles)
		if err != nil {
			http.Error(w, err.Error(), 403)
			return
		}
		err = checkUserHasRole(user, []string{os.Getenv("ADMIN_ROLE")})
		if err == nil {
			user.IsAdmin = true
		}
		// Check for url
		if user.URL != "" {
			requestURL := strings.Split(r.Host, ":")[0] + r.URL.EscapedPath()
			if user.URL != requestURL {
				http.Error(w, "token restricted to url: "+user.URL, 401)
				return
			}
		}
		// Check for method
		if user.ReadOnly && r.Method != http.MethodGet {
			http.Error(w, "token is read only", 403)
			return
		}
		ctx := context.WithValue(r.Context(), ContextData, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(roleChecker)
}

// HandleLogout remove the user from the cookie store
func (m Manager) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Delete the auth cookie
	c := http.Cookie{
		Name:   authTokenKey,
		Domain: m.Hostname,
		MaxAge: -1,
	}
	http.SetCookie(w, &c)
	http.Redirect(w, r, os.Getenv("LOGOUT_URL"), http.StatusTemporaryRedirect)
}

// WhoAmI returns the user data
func WhoAmI() http.Handler {
	whoAmI := func(w http.ResponseWriter, r *http.Request) {
		user, err := GetTokenData(r)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(user)
	}
	return http.HandlerFunc(whoAmI)
}

// checkUserHasRole checks if the user has the required role
func checkUserHasRole(user TokenData, allowedRoles []string) error {
	for _, allowedRole := range allowedRoles {
		if allowedRole == "*" {
			return nil
		}
		for _, userRole := range user.Roles {
			if userRole != "" && (userRole == allowedRole) {
				return nil
			}
		}
	}
	return fmt.Errorf("no user role among %v is in allowed roles (%v)", user.Roles, allowedRoles)
}

//GetShareToken gets a share token for a given ressource
func GetShareToken(w http.ResponseWriter, r *http.Request) {
	user, err := GetTokenData(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var wantedToken struct {
		Sharedfor string `json:"sharedfor"`
		URL       string `json:"url"`
		Lifespan  int    `json:"lifespan"`
		ReadOnly  bool   `json:"readonly,omitempty"`
	}
	err = json.NewDecoder(r.Body).Decode(&wantedToken)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if wantedToken.URL == "" {
		http.Error(w, "url cannot be empty", 400)
		return
	}
	user.Login = user.Login + "_share_for_" + wantedToken.Sharedfor
	user.URL = wantedToken.URL
	user.ReadOnly = wantedToken.ReadOnly
	user.SharingUserLogin = wantedToken.Sharedfor
	token, err := tokens.Manager.CreateToken(user, time.Now().Add(time.Hour*time.Duration(24*wantedToken.Lifespan)))
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	fmt.Fprintf(w, token)
}

// GetTokenData gets an user from a request
func GetTokenData(r *http.Request) (TokenData, error) {
	user, ok := r.Context().Value(ContextData).(TokenData)
	if !ok {
		return user, errors.New("user could not be got from context")
	}
	return user, nil
}
