package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nicolaspernoud/vestibule/pkg/common"
	"github.com/nicolaspernoud/vestibule/pkg/log"
	"github.com/nicolaspernoud/vestibule/pkg/tokens"
	"golang.org/x/crypto/bcrypt"
)

var (
	//UsersFile is the file containing the users
	UsersFile     = "./configs/users.json"
	tokenLifetime time.Duration
	cachedUsers   []User
)

func setTokenLifetime() time.Duration {
	days := 1
	i, err := strconv.Atoi(os.Getenv("INMEMORY_TOKEN_LIFE_DAYS"))
	if err == nil && i >= 1 && i <= 10000 {
		days = i
	}
	return time.Duration(days*24) * time.Hour
}

func refreshCache() {
	err := common.Load(UsersFile, &cachedUsers)
	if err != nil {
		log.Logger.Fatalln("could not load users")
	}
}

func init() {
	tokenLifetime = setTokenLifetime()
}

// HandleInMemoryLogin validate the username and password provided in the function body against a local file and return a token if the user is found
func (m Manager) HandleInMemoryLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var sentUser User
	err := json.NewDecoder(r.Body).Decode(&sentUser)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	// Try to match the user with an user in the database
	user, err := MatchUser(sentUser)
	if err != nil {
		http.Error(w, err.Error(), 403)
		log.Logger.Printf("| %v | Login failure | %v | %v", sentUser.Login, r.RemoteAddr, log.GetCityAndCountryFromRequest(r))
		return
	}
	// Store the user in cookie
	// Store only the relevant info
	// Generate
	xsrfToken, err := common.GenerateRandomString(16)
	if err != nil {
		http.Error(w, "error generating XSRF Token", 500)
		return
	}
	tokenData := TokenData{User: User{ID: user.ID, Login: user.Login, Email: user.Email, Roles: user.Roles}, XSRFToken: xsrfToken}
	tokens.Manager.StoreData(tokenData, m.Hostname, authTokenKey, tokenLifetime, w)
	// Log the connexion
	log.Logger.Printf("| %v (%v %v) | Login success | %v | %v", user.Login, user.Name, user.Surname, r.RemoteAddr, log.GetCityAndCountryFromRequest(r))
}

// ByID implements sort.Interface for []User based on the ID field
type ByID []User

func (a ByID) Len() int           { return len(a) }
func (a ByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByID) Less(i, j int) bool { return a[i].ID < a[j].ID }

// ProcessUsers processes users regarding of HTTP method
func ProcessUsers(w http.ResponseWriter, req *http.Request) {
	switch method := req.Method; method {
	case "GET":
		SendUsers(w, req)
	case "POST":
		AddUser(w, req)
		refreshCache()
	case "DELETE":
		DeleteUser(w, req)
		refreshCache()
	default:
		http.Error(w, "method not allowed", 400)
	}
}

// SendUsers send users as response from an http requests
func SendUsers(w http.ResponseWriter, req *http.Request) {
	var users []User
	err := common.Load(UsersFile, &users)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	json.NewEncoder(w).Encode(users)
}

// AddUser adds an user
func AddUser(w http.ResponseWriter, req *http.Request) {
	var users []User
	err := common.Load(UsersFile, &users)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if req.Body == nil {
		http.Error(w, "please send a request body", 400)
		return
	}
	var newUser User
	err = json.NewDecoder(req.Body).Decode(&newUser)
	if _, ok := err.(*json.UnmarshalTypeError); !ok && err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	// Encrypt the password with bcrypt
	if newUser.Password == "" && newUser.PasswordHash == "" {
		http.Error(w, "passwords cannot be blank", 400)
		return
	}
	if newUser.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		newUser.PasswordHash = string(hash)
		newUser.Password = ""
	}
	// Add the user only if the id doesn't exists yet
	isNew := true
	for idx, val := range users {
		if val.ID == newUser.ID {
			users[idx] = newUser
			isNew = false
		} else if val.Login == newUser.Login { // Check for already existing login
			http.Error(w, "login already exists", 400)
			return
		}
	}
	if isNew {
		users = append(users, newUser)
		sort.Sort(ByID(users))
	}
	err = common.Save(UsersFile, &users)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	SendUsers(w, req)
}

// DeleteUser deletes an user
func DeleteUser(w http.ResponseWriter, req *http.Request) {
	var users []User
	err := common.Load(UsersFile, &users)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	pathElements := strings.Split(req.URL.Path, "/")
	idx, err := strconv.Atoi(pathElements[len(pathElements)-1])
	if err != nil {
		http.Error(w, "please provide an user index", 400)
		return
	}
	// Recreate the user list without the deleted user
	newUsers := users[:0]
	for _, user := range users {
		id, err := strconv.Atoi(user.ID)
		if err == nil && id != idx {
			newUsers = append(newUsers, user)
		}
	}
	err = common.Save(UsersFile, &newUsers)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	SendUsers(w, req)
}

// MatchUser attempt to find the given user against users in configuration file
func MatchUser(sentUser User) (User, error) {
	if len(cachedUsers) == 0 {
		refreshCache()
	}
	var emptyUser User
	for _, user := range cachedUsers {
		if user.Login == sentUser.Login {
			notFound := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(sentUser.Password))
			if notFound == nil {
				return user, nil
			}
		}
	}
	return emptyUser, errors.New("user not found")
}
