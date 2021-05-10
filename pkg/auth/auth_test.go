package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"

	b64 "encoding/base64"

	"github.com/nicolaspernoud/vestibule/pkg/common"
	"github.com/nicolaspernoud/vestibule/pkg/tester"
)

var noH map[string]string

func TestCheckUserHasRole(t *testing.T) {
	type args struct {
		user         TokenData
		allowedRoles []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"has_all_roles", args{user: TokenData{User: User{Roles: []string{"role01", "role02"}}}, allowedRoles: []string{"role01", "role02"}}, false},
		{"has_one_role", args{user: TokenData{User: User{Roles: []string{"role01", "role03"}}}, allowedRoles: []string{"role01", "role02"}}, false},
		{"has_not_role", args{user: TokenData{User: User{Roles: []string{"role03", "role04"}}}, allowedRoles: []string{"role01", "role02"}}, true},
		{"user_roles_are_empty", args{user: TokenData{User: User{Roles: []string{}}}, allowedRoles: []string{"role01", "role02"}}, true},
		{"allowed_roles_are_empty", args{user: TokenData{User: User{Roles: []string{"role01", "role02"}}}, allowedRoles: []string{}}, true},
		{"all_roles_are_empty", args{user: TokenData{User: User{Roles: []string{}}}, allowedRoles: []string{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkUserHasRole(tt.args.user, tt.args.allowedRoles); (err != nil) != tt.wantErr {
				t.Errorf("checkUserHasRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddUser(t *testing.T) {
	UsersFile = writeUsers()
	defer os.Remove(UsersFile)

	handler := http.HandlerFunc(AddUser)
	// Alter the password of the admin user, must create an hash
	tester.DoRequestOnHandler(t, handler, "POST", "/", noH, `{"id":"1","login":"admin","password": "password"}`, http.StatusOK, `[{"id":"1","login":"admin","memberOf":null,"passwordHash":"$2a`)
	// Test that altering an user without altering the password keep the password hash
	tester.DoRequestOnHandler(t, handler, "POST", "/", noH, `{"id":"1","login":"admin_altered"}`, http.StatusOK, `[{"id":"1","login":"admin_altered","memberOf":null,"passwordHash":"$2a`)
	// Add a new user with a password, must pass
	tester.DoRequestOnHandler(t, handler, "POST", "/", noH, `{"id":"3","login":"user3","password": "password_user3"}`, http.StatusOK, `[{"id":"1","login":"admin`)
	// Add a new user with no password, must fail
	tester.DoRequestOnHandler(t, handler, "POST", "/", noH, `{"id":"4","login":"user4","password": ""}`, http.StatusBadRequest, `password cannot be empty`)
}

func TestMatchUser(t *testing.T) {
	UsersFile = "../../configs/users.json"
	existingUser := User{ID: "2", Login: "user", Roles: []string{"USERS"}, PasswordHash: "$2a$10$PgiAoLxZhgNtr7kRK/DH5ezwT./7vRkWqFNEtJD1670z3Zf60HqgG"}
	veryLongString, _ := common.GenerateRandomString(10000)
	specialCharString := "\""

	type args struct {
		sentUser User
	}
	tests := []struct {
		name    string
		args    args
		want    User
		wantErr bool
	}{
		{"user_exists", args{User{Login: "user", Password: "password"}}, existingUser, false},
		{"user_does_not_exists", args{User{Login: "notuser", Password: "password"}}, User{}, true},
		{"user_does_not_exists_and_wrong_password", args{User{Login: "notuser", Password: "wrongpassword"}}, User{}, true},
		{"wrong_password", args{User{Login: "user", Password: "wrongpassword"}}, User{}, true},
		{"no_password", args{User{Login: "user", Password: ""}}, User{}, true},
		{"empty_user", args{User{Login: "", Password: "password"}}, User{}, true},
		{"empty_user_and_password", args{User{Login: "", Password: ""}}, User{}, true},
		{"very_long_string_as_user", args{User{Login: veryLongString, Password: "password"}}, User{}, true},
		{"very_long_string_as_password", args{User{Login: "user", Password: veryLongString}}, User{}, true},
		{"very_long_string_as_user_and_password", args{User{Login: veryLongString, Password: veryLongString}}, User{}, true},
		{"special_char_string_as_user", args{User{Login: specialCharString, Password: "password"}}, User{}, true},
		{"special_char_string_as_password", args{User{Login: "user", Password: specialCharString}}, User{}, true},
		{"special_char_string_as_user_and_password", args{User{Login: specialCharString, Password: specialCharString}}, User{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchUser(tt.args.sentUser)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MatchUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func writeUsers() (name string) {
	users := []*User{
		{ID: "1", Login: "admin"},
		{ID: "2", Login: "user"},
	}
	f, err := ioutil.TempFile("", "users")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(users)
	if err != nil {
		panic(err)
	}
	return f.Name()
}

func TestIsWebdav(t *testing.T) {
	type args struct {
		ua string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"is_exact", args{ua: "Microsoft-WebDAV"}, true},
		{"contains", args{ua: "Contains-Microsoft-WebDAV-"}, true},
		{"is_not", args{ua: "Microsoft-Other"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isWebdav(tt.args.ua); got != tt.want {
				t.Errorf("isWebdav() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUserDirectly(t *testing.T) {
	UsersFile = `../../configs/users.json`
	sentUser := User{Login: "user", Password: "password", Roles: []string{"USERS"}}
	sentAdmin := User{Login: "admin", Password: "password", Roles: []string{"ADMINS"}}
	wrongUser := User{Login: "user", Password: "wrong_password"}
	authFromUser := func(user User) string {
		data := user.Login + ":" + user.Password
		return "Basic " + b64.StdEncoding.EncodeToString([]byte(data))
	}
	type args struct {
		authorizationHeader string
	}
	tests := []struct {
		name    string
		args    args
		want    TokenData
		wantErr bool
	}{
		{"user", args{authorizationHeader: authFromUser(sentUser)}, TokenData{User: sentUser}, false},
		{"admin", args{authorizationHeader: authFromUser(sentAdmin)}, TokenData{User: sentAdmin}, false},
		{"wrong_user", args{authorizationHeader: authFromUser(wrongUser)}, TokenData{User: wrongUser}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getUserDirectly(tt.args.authorizationHeader)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUserDirectly() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !(got.Login == tt.want.Login && got.Roles[0] == tt.want.Roles[0]) {
				t.Errorf("getUserDirectly() = %v, want %v", got, tt.want)
			}
		})
	}
}
