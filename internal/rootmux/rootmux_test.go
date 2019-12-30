package rootmux

import (
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/nicolaspernoud/vestibule/pkg/auth"
	"github.com/nicolaspernoud/vestibule/pkg/jwt"
	"github.com/nicolaspernoud/vestibule/pkg/tester"

	"github.com/nicolaspernoud/vestibule/internal/mocks"
)

var (
	initialAppsBuff, _     = ioutil.ReadFile("../../configs/apps.json")
	reg, _                 = regexp.Compile("[\n \t]+")
	initialApps            = reg.ReplaceAllString(string(initialAppsBuff), "")
	newApp                 = "{\"id\":4,\"host\":\"test\",\"isProxy\":false,\"forwardTo\":\"\",\"serve\":\"test\"}"
	updatedAppsWithSchemes = strings.Replace(initialApps, "api.vestibule.", "http://api.vestibule.", 1)
	initialUsersBuff, _    = ioutil.ReadFile("../../configs/users.json")
	initialUsers           = reg.ReplaceAllString(string(initialUsersBuff), "")
	newUser                = `{"id":3,"login":"new_user","memberOf":["USERS","OTHERS"],"password":"test"}`
)

func init() {
	jwt.Init(true)
}

func TestAll(t *testing.T) {
	// Set the users file
	auth.UsersFile = "../../configs/users.json"
	// Create the mock OAuth2 server
	oAuth2Server := httptest.NewServer(mocks.CreateMockOAuth2())
	defer oAuth2Server.Close()
	// Create the mock API server
	go http.ListenAndServe(":8091", mocks.CreateMockAPI())
	// Set the constants with environment variables
	os.Setenv("HOSTNAME", "vestibule.io")
	os.Setenv("ADMIN_ROLE", "ADMINS")
	os.Setenv("COMMON_ROLE", "OTHERS")
	os.Setenv("CLIENT_ID", "clientid")
	os.Setenv("CLIENT_SECRET", "clientsecret")
	os.Setenv("TOKEN_URL", oAuth2Server.URL+"/token")
	os.Setenv("USERINFO_URL", oAuth2Server.URL+"/userinfo")
	os.Setenv("LOGOUT_URL", oAuth2Server.URL+"/logout")
	// Set up testers
	os.Setenv("AUTH_URL", oAuth2Server.URL+"/auth-wrong-state") // Set the server to access failing OAuth2 endpoints
	oauth2Tests := createOauth2Tests(t)
	os.Setenv("AUTH_URL", oAuth2Server.URL+"/auth") // Set the server to access the correct OAuth2Endpoint
	unloggedTests := createUnLoggedTests(t)
	userTests := createUserTests(t)
	os.Setenv("USERINFO_URL", oAuth2Server.URL+"/admininfo")
	adminTests := createAdminTests(t)
	// RUN THE TESTS CONCURRENTLY
	var wg sync.WaitGroup
	functions := []func(wg *sync.WaitGroup){oauth2Tests, unloggedTests, userTests, adminTests}
	for _, f := range functions {
		wg.Add(1)
		go f(&wg)
	}
	wg.Wait()

}

/**
SECURITY TESTS (this tests are to check that the security protections works)
**/
func createOauth2Tests(t *testing.T) func(wg *sync.WaitGroup) {
	// Create the tester
	ts, do := createTester(t)
	return func(wg *sync.WaitGroup) {
		defer ts.Close() // Close the tester
		defer wg.Done()
		// Try to login (must fail)
		do("GET", "/OAuth2Login", "", "", 500, "invalid oauth state")
	}
}

/**
UNLOGGED USER TESTS (this tests are to check that the security protections works)
**/
func createUnLoggedTests(t *testing.T) func(wg *sync.WaitGroup) {
	// Create the tester
	ts, do := createTester(t)
	return func(wg *sync.WaitGroup) {
		defer ts.Close() // Close the tester
		defer wg.Done()
		// Try to get the apps (must fail)
		do("GET", "/api/admin/apps", "", "", 401, "error extracting JWT: no token found")
		// Try to create an app (must fail)
		do("POST", "/api/admin/apps/", "", newApp, 401, "error extracting JWT: no token found")
		// Try to delete an app (must fail)
		do("DELETE", "/api/admin/apps/4", "", "", 401, "error extracting JWT: no token found")
		// Try to get the users (must fail)
		do("GET", "/api/admin/users/", "", "", 401, "error extracting JWT: no token found")
		// Try to create an user (must fail)
		do("POST", "/api/admin/users/", "", newUser, 401, "error extracting JWT: no token found")
		// Try to delete an user (must fail)
		do("DELETE", "/api/admin/users/0", "", "", 401, "error extracting JWT: no token found")
		// Try to access the apps list (must fail)
		do("GET", "/api/common/apps", "", "", 401, "error extracting JWT: no token found")
		// Try to access an app (must fail)
		do("GET", "api.vestibule.io", "", "", 401, "error extracting JWT: no token found")
		// Try to access the main page (must pass)
		do("GET", "/", "", "", 200, "<!DOCTYPE html>")
		// Try to get the user informations (must fail)
		do("GET", "/api/common/WhoAmI", "", "", 401, "error extracting JWT: no token found")
		// Do a in memory login with an unknown user
		do("POST", "/Login", "", `{"login": "unknownuser","password": "password"}`, http.StatusForbidden, `user not found`)
		// Do a in memory login with a known user but bad password
		do("POST", "/Login", "", `{"login": "admin","password": "badpassword"}`, http.StatusForbidden, `user not found`)
	}
}

/**
USER TESTS (this tests are to check that a normally logged user can access the apps that is allowed to and only that)
**/
func createUserTests(t *testing.T) func(wg *sync.WaitGroup) {
	// Create the tester
	ts, do := createTester(t)
	return func(wg *sync.WaitGroup) {
		defer ts.Close() // Close the tester
		defer wg.Done()
		tests := func() {
			// Try to get the apps (must fail)
			do("GET", "/api/admin/apps", "", "", 403, "no user role among")
			// Try to create an app (must fail)
			do("POST", "/api/admin/apps/", "", newApp, 403, "no user role among")
			// Try to delete an app (must fail)
			do("DELETE", "/api/admin/apps/4", "", "", 403, "no user role among")
			// Try to get the users (must fail)
			do("GET", "/api/admin/users/", "", "", 403, "no user role among")
			// Try to create an user (must fail)
			do("POST", "/api/admin/users/", "", newUser, 403, "no user role among")
			// Try to delete an user (must fail)
			do("DELETE", "/api/admin/users/0", "", "", 403, "no user role among")
			// Try to access the apps list (must pass)
			do("GET", "/api/common/apps", "", "", 200, "[{\"id\":1,\"isProxy\"")
			// Try to access a forbidden app (must fail)
			do("GET", "external.vestibule.io", "", "", 403, "no user role among")
			// Try to access the main page (must pass)
			do("GET", "/", "", "", 200, "<!DOCTYPE html>")
			// Try to access an authorized app (must pass)
			do("GET", "api.vestibule.io", "", "", 200, "{")
			// Try to get the user informations (must pass)
			do("GET", "/api/common/WhoAmI", "", "", 200, `{"id":`)
		}
		// Try to login with OAuth2 (must pass)
		do("GET", "/OAuth2Login", "", "", 200, "<!DOCTYPE html>")
		// Run the tests
		tests()
		// Try to logout (must pass)
		do("GET", "/Logout", "", "", 200, "Logout OK")
		// Try to access an authorized app after logout (must fail)
		do("GET", "api.vestibule.io", "", "", 401, "error extracting JWT: no token found")
		// Do a in memory login with an known user
		do("POST", "/Login", "", `{"login": "user","password": "password"}`, 200, "<!DOCTYPE html>")
		// Run the tests
		tests()
		// Try to logout (must pass)
		do("GET", "/Logout", "", "", 200, "Logout OK")
		// Try to access an authorized app after logout (must fail)
		do("GET", "api.vestibule.io", "", "", 401, "error extracting JWT: no token found")
	}
}

/**
ADMIN TESTS (this tests are to check that an administrator can alter the apps)
**/
func createAdminTests(t *testing.T) func(wg *sync.WaitGroup) {
	// Create the tester
	ts, do := createTester(t)
	return func(wg *sync.WaitGroup) {
		defer ts.Close() // Close the tester
		defer wg.Done()
		tests := func() {
			// Try to get the apps (must pass)
			do("GET", "/api/admin/apps/", "", "", 200, "[{\"id\":1")
			// Try to create an app (must pass)
			do("POST", "/api/admin/apps/", "", newApp, 200, "[{\"id\":1")
			// Try to delete an app (must pass)
			do("DELETE", "/api/admin/apps/4", "", "", 200, "[{\"id\":1")
			// Try to get the users (must pass)
			do("GET", "/api/admin/users/", "", "", 200, "[{\"id\":1,")
			// Try to create an user (must pass)
			do("POST", "/api/admin/users/", "", newUser, 200, "[{\"id\":1,")
			// Try to delete an user (must pass)
			do("DELETE", "/api/admin/users/3", "", "", 200, "[{\"id\":1,")
			// Try to get the user informations (must pass)
			do("GET", "/api/common/WhoAmI", "", "", 200, `{"id":`)
		}
		// Try to login (must pass)
		do("GET", "/OAuth2Login", "", "", 200, "<!DOCTYPE html>")
		tests()
		// Try to logout (must pass)
		do("GET", "/Logout", "", "", 200, "Logout OK")
		// Try to get the apps again (must fail)
		do("GET", "/api/admin/apps", "", "", 401, "error extracting JWT: no token found")
		// Do a in memory login with an known admin
		do("POST", "/Login", "", `{"login": "admin","password": "password"}`, 200, "<!DOCTYPE html>")
		tests()
		// Try to logout (must pass)
		do("GET", "/Logout", "", "", 200, "Logout OK")
		// Try to get the apps again (must fail)
		do("GET", "/api/admin/apps", "", "", 401, "error extracting JWT: no token found")
	}
}

func createTester(t *testing.T) (*httptest.Server, func(method string, url string, authHeader string, payload string, expectedStatus int, expectedBody string)) {
	// Create the server
	mux := CreateRootMux(1443, "./testdata/apps.json", "../../web")
	ts := httptest.NewServer(mux.Mux)
	url, _ := url.Parse(ts.URL)
	port := url.Port()
	mux.Manager.Config.RedirectURL = "http://" + os.Getenv("HOSTNAME") + ":" + port + "/OAuth2Callback"
	mux.Manager.Hostname = "http://" + os.Getenv("HOSTNAME") + ":" + port
	// Create the cookie jar
	jar, _ := cookiejar.New(nil)
	// wrap the testing function
	return ts, tester.CreateServerTester(t, port, os.Getenv("HOSTNAME"), jar)
}
