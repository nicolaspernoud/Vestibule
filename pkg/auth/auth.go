package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nicolaspernoud/vestibule/pkg/common"
	"github.com/nicolaspernoud/vestibule/pkg/log"
	"github.com/nicolaspernoud/vestibule/pkg/tokens"
)

type key int

const (
	authTokenKey string = "auth_token"
	// ContextData is the user
	ContextData key = 0
)

type OpenIDConfiguration struct {
	ResponseTypesSupported           []string `json:"response_types_supported"`
	RequestParameterSupported        bool     `json:"request_parameter_supported"`
	RequestURIParameterSupported     bool     `json:"request_uri_parameter_supported"`
	JwksURI                          string   `json:"jwks_uri"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	RegistrationEndpoint             string   `json:"registration_endpoint"`
	Issuer                           string   `json:"issuer"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	UserinfoEndpoint                 string   `json:"userinfo_endpoint"`
}

// InitEnv will initialize environment variables from the issuer environment variable
func InitIdPEnv(idpUrl string) {
	// Perform a request on the .well-known/oidc-configuration endpoint
	r, err := http.Get(idpUrl + "/.well-known/openid-configuration")
	if err != nil {
		log.Logger.Fatalf("Could not read IdP configuration, exiting...")
	}
	defer r.Body.Close()
	// Unmarshall the response into a struct
	var o = OpenIDConfiguration{}
	json.NewDecoder(r.Body).Decode(&o)
	// Init the other env variables from this struct
	if _, e := os.LookupEnv("AUTH_URL"); !e {
		os.Setenv("AUTH_URL", o.AuthorizationEndpoint)
	}
	if _, e := os.LookupEnv("TOKEN_URL"); !e {
		os.Setenv("TOKEN_URL", o.TokenEndpoint)
	}
	if _, e := os.LookupEnv("USERINFO_URL"); !e {
		os.Setenv("USERINFO_URL", o.UserinfoEndpoint)
	}
}

var (
	// AdminRole represents the role reserved for admins
	AdminRole = common.StringValueFromEnv("ADMIN_ROLE", "ADMINS")
	hostname  = common.StringValueFromEnv("HOSTNAME", "vestibule.127.0.0.1.nip.io")
)

// User represents a logged in user
type User struct {
	ID           string   `json:"id,omitempty"`
	Login        string   `json:"login"`
	DisplayName  string   `json:"displayName,omitempty"`
	Email        string   `json:"email,omitempty"`
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
		checkXSRF, err := tokens.ExtractAndValidateToken(r, authTokenKey, &user, checkXSRF)
		// Handle WebDav authentication
		if err != nil && isWebdav(r.UserAgent()) {
			// Test if the user password is directly given in the request, if so populate the user
			user, err = getUserDirectly(r.Header.Get("Authorization"))
			if err != nil {
				w.Header().Set("WWW-Authenticate", `Basic realm="server"`)
				http.Error(w, "webdav client authentication", http.StatusUnauthorized)
				return
			}
		}
		if err != nil {
			// Handle CORS preflight requests
			if r.Method == "OPTIONS" {
				return
			}
			// Default to redirect to authentication
			redirectTo := hostname
			_, port, perr := net.SplitHostPort(r.Host)
			if perr == nil {
				redirectTo += ":" + port
			}
			// Write the requested url in a cookie
			if r.Host != redirectTo && r.URL.Path != "/favicon.ico" {
				cookie := http.Cookie{Name: "redirectAfterLogin", Path: "/", Domain: hostname, Value: r.Host + r.URL.Path + "?" + r.URL.RawQuery, MaxAge: 60, Secure: true, HttpOnly: false, SameSite: http.SameSiteLaxMode}
				http.SetCookie(w, &cookie)
			}
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusUnauthorized)
			responseContent := fmt.Sprintf("error extracting token: %v<meta http-equiv=\"Refresh\" content=\"0; url=https://%v#login\"/>", err.Error(), redirectTo)
			fmt.Fprint(w, responseContent)
			return

		}
		// Check XSRF Token
		if checkXSRF && r.Header.Get("XSRF-TOKEN") != user.XSRFToken {
			http.Error(w, "XSRF protection triggered", http.StatusUnauthorized)
			return
		}
		err = checkUserHasRole(user, allowedRoles)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		err = checkUserHasRole(user, []string{AdminRole})
		if err == nil {
			user.IsAdmin = true
		}
		// Check for url
		if user.URL != "" {
			requestURL := strings.Split(r.Host, ":")[0] + r.URL.EscapedPath()
			if user.URL != requestURL {
				http.Error(w, "token restricted to url: "+user.URL, http.StatusUnauthorized)
				return
			}
		}
		// Check for method
		if user.ReadOnly && r.Method != http.MethodGet {
			http.Error(w, "token is read only", http.StatusForbidden)
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
			http.Error(w, err.Error(), http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if wantedToken.URL == "" {
		http.Error(w, "url cannot be empty", http.StatusBadRequest)
		return
	}
	user.Login = user.Login + "_share_for_" + wantedToken.Sharedfor
	user.URL = wantedToken.URL
	user.ReadOnly = wantedToken.ReadOnly
	user.SharingUserLogin = wantedToken.Sharedfor
	token, err := tokens.CreateToken(user, time.Now().Add(time.Hour*time.Duration(24*wantedToken.Lifespan)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, token)
}

// GetTokenData gets an user from a request
func GetTokenData(r *http.Request) (TokenData, error) {
	user, ok := r.Context().Value(ContextData).(TokenData)
	if !ok {
		return user, errors.New("user could not be got from context")
	}
	return user, nil
}

// isWebdav works out if an user agent is a webdav user agent
func isWebdav(ua string) bool {
	for _, a := range []string{"vfs", "Microsoft-WebDAV", "Konqueror", "LibreOffice", "Rei.Fs.WebDAV", "Documents"} {
		if strings.Contains(ua, a) {
			return true
		}
	}
	return false
}

// getUserDirectly directly checks if an user is allowed to connect
func getUserDirectly(authorizationHeader string) (TokenData, error) {
	authHeader := strings.Split(authorizationHeader, " ")
	var user User
	if authHeader[0] == "Basic" && len(authHeader) == 2 {
		decoded, err := base64.StdEncoding.DecodeString(authHeader[1])
		if err == nil {
			auth := strings.Split(string(decoded), ":")
			sentUser := User{Login: auth[0], Password: auth[1]}
			foundUser, err := MatchUser(sentUser)
			if err == nil {
				return (TokenData{User: foundUser}), nil
			}
		}
	}
	return TokenData{User: user}, errors.New("could not retrieve user directly from basic auth header")
}
