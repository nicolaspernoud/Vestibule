package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/nicolaspernoud/vestibule/pkg/jwt"
)

const (
	authTokenKey string = "auth_token"
)

// User represents a logged in user
type User struct {
	ID           int      `json:"id,omitempty"`
	Login        string   `json:"login"`
	DisplayName  string   `json:"displayName,omitempty"`
	Roles        []string `json:"memberOf"`
	IsAdmin      bool     `json:"isAdmin,omitempty"`
	Name         string   `json:"name,omitempty"`
	Surname      string   `json:"surname,omitempty"`
	PasswordHash string   `json:"passwordHash,omitempty"`
	Password     string   `json:"password,omitempty"`
}

// ValidateJWTAndRolesMiddleware validates that the token is valid and that the user has the correct roles
func ValidateJWTAndRolesMiddleware(next http.Handler, allowedRoles []string) http.Handler {
	roleChecker := func(w http.ResponseWriter, r *http.Request) {
		user, err := GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err = checkUserHasRole(user, allowedRoles)
		if err != nil {
			http.Error(w, err.Error(), 403)
			return
		}
		next.ServeHTTP(w, r)
	}
	return jwt.ValidateJWTMiddleware(http.HandlerFunc(roleChecker), authTokenKey)
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
func WhoAmI(w http.ResponseWriter, r *http.Request) {
	user, err := GetUser(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	json.NewEncoder(w).Encode(user)
}

// checkUserHasRole checks if the user has the required role
func checkUserHasRole(user User, allowedRoles []string) error {
	for _, allowedRole := range allowedRoles {
		for _, userRole := range user.Roles {
			if userRole != "" && (userRole == allowedRole) {
				return nil
			}
		}
	}
	return fmt.Errorf("no user role among %v is in allowed roles (%v)", user.Roles, allowedRoles)
}

// GetUser gets an user from a request
func GetUser(r *http.Request) (User, error) {
	var user User
	data := r.Context().Value(jwt.ContextData)
	dataStr, err := json.Marshal(data)
	if err != nil {
		return user, errors.New("user could not be recovered from context")
	}
	err = json.Unmarshal([]byte(dataStr), &user)
	if err != nil {
		return user, errors.New("user could not be unmarshaled from json")
	}
	err = checkUserHasRole(user, []string{os.Getenv("ADMIN_ROLE")})
	if err == nil {
		user.IsAdmin = true
	}
	return user, nil
}
