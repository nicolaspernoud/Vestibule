package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/nicolaspernoud/vestibule/pkg/common"
	"github.com/nicolaspernoud/vestibule/pkg/tester"
)

func Test_checkUserHasRole(t *testing.T) {
	type args struct {
		user         User
		allowedRoles []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"has_all_roles", args{user: User{Roles: []string{"role01", "role02"}}, allowedRoles: []string{"role01", "role02"}}, false},
		{"has_one_role", args{user: User{Roles: []string{"role01", "role03"}}, allowedRoles: []string{"role01", "role02"}}, false},
		{"has_not_role", args{user: User{Roles: []string{"role03", "role04"}}, allowedRoles: []string{"role01", "role02"}}, true},
		{"user_roles_are_empty", args{user: User{Roles: []string{}}, allowedRoles: []string{"role01", "role02"}}, true},
		{"allowed_roles_are_empty", args{user: User{Roles: []string{"role01", "role02"}}, allowedRoles: []string{}}, true},
		{"all_roles_are_empty", args{user: User{Roles: []string{}}, allowedRoles: []string{}}, true},
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
	tester.DoRequestOnHandler(t, handler, "POST", "/", "", `{"id":1,"login":"admin","password": "password"}`, http.StatusOK, `[{"id":1,"login":"admin"`)
	tester.DoRequestOnHandler(t, handler, "POST", "/", "", `{"id":1,"login":"admin","password": ""}`, http.StatusBadRequest, `passwords cannot be blank`)
}

func TestMatchUser(t *testing.T) {
	UsersFile = "../../configs/users.json"
	existingUser := User{ID: 2, Login: "user", Roles: []string{"USERS", "OTHERS"}, PasswordHash: "$2a$10$PgiAoLxZhgNtr7kRK/DH5ezwT./7vRkWqFNEtJD1670z3Zf60HqgG"}
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
		{ID: 1, Login: "admin", Password: "password"},
		{ID: 2, Login: "user", Password: "password"},
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
