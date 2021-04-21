package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nicolaspernoud/vestibule/pkg/common"
	"github.com/nicolaspernoud/vestibule/pkg/log"
	"github.com/nicolaspernoud/vestibule/pkg/tokens"
	"golang.org/x/oauth2"
)

const (
	oAuth2StateKey string = "oauth2_state"
)

// Manager exposes the handlers for OAuth2 endpoints
type Manager struct {
	Config      *oauth2.Config
	Hostname    string
	UserInfoURL string
}

// NewManager returns a new Manager according to environment variables
func NewManager() Manager {
	return Manager{Config: &oauth2.Config{
		RedirectURL:  os.Getenv("REDIRECT_URL"),
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"login", "memberOf", "displayName", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  os.Getenv("AUTH_URL"),
			TokenURL: os.Getenv("TOKEN_URL"),
		},
	},
		Hostname:    common.StringValueFromEnv("HOSTNAME", "vestibule.127.0.0.1.nip.io"),
		UserInfoURL: os.Getenv("USERINFO_URL"),
	}
}

// HandleOAuth2Login handles the OAuth2 login
func (m Manager) HandleOAuth2Login(w http.ResponseWriter, r *http.Request) {
	// Generate state and store it in cookie
	oauthStateString, err := common.GenerateRandomString(16)
	if err != nil {
		log.Logger.Fatalf("Error generating OAuth2 strate string :%v\n", err)
	}
	tokens.Manager.StoreData(oauthStateString, m.Hostname, oAuth2StateKey, 30*time.Second, w)
	url := m.Config.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleOAuth2Callback handles the OAuth2 Callback and get user info
func (m Manager) HandleOAuth2Callback() http.Handler {
	oauth2Handler := func(w http.ResponseWriter, r *http.Request) {
		// Recover state from tokens
		var oauthState string
		_, err := tokens.Manager.ExtractAndValidateToken(r, oAuth2StateKey, &oauthState, false)
		if err != nil {
			fmt.Println("Code exchange failed with ", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		// Check states match
		state := r.FormValue("state")
		if state != oauthState {
			fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthState, state)
			http.Error(w, "invalid oauth state", http.StatusInternalServerError)
			return
		}
		// Delete the state cookie
		c := http.Cookie{
			Name:   oAuth2StateKey,
			Domain: m.Hostname,
			MaxAge: -1,
		}
		http.SetCookie(w, &c)
		// Perform code exchange
		code := r.FormValue("code")
		token, err := m.Config.Exchange(context.Background(), code)
		if err != nil {
			fmt.Printf("Code exchange failed with '%s'\n", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		// Get user infos
		client := &http.Client{}
		req, _ := http.NewRequest("GET", m.UserInfoURL+"?access_token="+token.AccessToken, nil)
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
		response, err := client.Do(req)
		if err != nil || response.StatusCode == http.StatusBadRequest {
			fmt.Printf("User info failed with '%s'\n", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		// Get user
		var user User
		if response.Body == nil {
			http.Error(w, "no response body", http.StatusBadRequest)
			return
		}
		////////////////////////////////////////////////
		// UNCOMMENT THIS TO DEBUG USERINFO RESPONSE //
		/*readBody, err := ioutil.ReadAll(response.Body)
		  if err != nil {
		      panic(err)
		  }
		  newBody := ioutil.NopCloser(bytes.NewBuffer(readBody))
		  response.Body = newBody
		  if string(readBody) != "" {
		      fmt.Printf("BODY : %q", readBody)
		  }*/
		////////////////////////////////////////////////
		err = json.NewDecoder(response.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Trim the user roles in case they come from LDAP
		for key, role := range user.Roles {
			user.Roles[key] = strings.TrimPrefix(strings.Split(role, ",")[0], "CN=")
		}
		// Store the user in cookie
		// Generate
		xsrfToken, err := common.GenerateRandomString(16)
		if err != nil {
			http.Error(w, "error generating XSRF Token", http.StatusInternalServerError)
			return
		}
		tokenData := TokenData{User: user, XSRFToken: xsrfToken}
		tokens.Manager.StoreData(tokenData, m.Hostname, authTokenKey, 24*time.Hour, w)
		// Log the connexion
		log.Logger.Printf("| %v (%v %v) | Login success | %v | %v", user.Login, user.Name, user.Surname, req.RemoteAddr, log.GetCityAndCountryFromRequest(req))
		// Redirect
		http.Redirect(w, r, "/", http.StatusFound)
	}
	return http.HandlerFunc(oauth2Handler)
}
