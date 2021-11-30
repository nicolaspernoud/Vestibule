package rootmux

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/nicolaspernoud/vestibule/pkg/auth"
	"github.com/nicolaspernoud/vestibule/pkg/tester"
	"github.com/nicolaspernoud/vestibule/pkg/tokens"

	b64 "encoding/base64"

	"github.com/nicolaspernoud/vestibule/internal/mocks"
)

var (
	newApp  = "{\"id\":4,\"host\":\"test\",\"isProxy\":false,\"forwardTo\":\"\",\"serve\":\"test\"}"
	newUser = `{"id":"3","login":"new_user","memberOf":["USERS"],"password":"test"}`
	newDav  = `{"id": 5,"host":"writableadmindav.vestibule.io","root":"./testdata/data","secured":true,"writable":true,"roles":["ADMINS"]}`
	noH     map[string]string
)

func init() {
	tokens.Init("testdata/tokenskey.json", true)
}

func TestAll(t *testing.T) {
	// Set the users file
	auth.UsersFile = "testdata/users.json"
	// Create the mock OAuth2 server
	oAuth2Server := httptest.NewServer(mocks.CreateMockOAuth2())
	defer oAuth2Server.Close()
	// Create the mock API server
	go http.ListenAndServe(":8091", mocks.CreateMockAPI())
	// Set the constants with environment variables
	os.Setenv("HOSTNAME", "vestibule.io")
	os.Setenv("ADMIN_ROLE", "ADMINS")
	os.Setenv("CLIENT_ID", "clientid")
	os.Setenv("CLIENT_SECRET", "clientsecret")
	os.Setenv("TOKEN_URL", oAuth2Server.URL+"/token")
	os.Setenv("USERINFO_URL", oAuth2Server.URL+"/userinfo")
	os.Setenv("LOGOUT_URL", oAuth2Server.URL+"/logout")
	// Set up testers
	os.Setenv("AUTH_URL", oAuth2Server.URL+"/authorize-wrong-state") // Set the server to access failing OAuth2 endpoints
	oauth2Tests := createOauth2Tests(t)
	os.Setenv("AUTH_URL", oAuth2Server.URL+"/authorize") // Set the server to access the correct OAuth2Endpoint
	unloggedTests := createUnLoggedTests(t)
	userTests := createUserTests(t)
	os.Setenv("USERINFO_URL", oAuth2Server.URL+"/admininfo")
	adminTests := createAdminTests(t)
	directWebdavTests := createDirectWebdavTests(t)
	// RUN THE TESTS CONCURRENTLY
	var wg sync.WaitGroup
	functions := []func(wg *sync.WaitGroup){oauth2Tests, unloggedTests, userTests, adminTests, directWebdavTests}
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
	ts, do, _ := createTester(t)
	return func(wg *sync.WaitGroup) {
		defer ts.Close() // Close the tester
		defer wg.Done()
		// Try to login (must fail)
		do("GET", "/OAuth2Login", noH, "", http.StatusInternalServerError, "invalid oauth state")
	}
}

/**
UNLOGGED USER TESTS (this tests are to check that the security protections works)
**/
func createUnLoggedTests(t *testing.T) func(wg *sync.WaitGroup) {
	// Create the tester
	ts, do, _ := createTester(t)
	return func(wg *sync.WaitGroup) {
		defer ts.Close() // Close the tester
		defer wg.Done()
		// Try to get the apps (must fail)
		do("GET", "/api/admin/apps", noH, "", http.StatusUnauthorized, "error extracting token")
		// Try to create an app (must fail)
		do("POST", "/api/admin/apps/", noH, newApp, http.StatusUnauthorized, "error extracting token")
		// Try to delete an app (must fail)
		do("DELETE", "/api/admin/apps/4", noH, "", http.StatusUnauthorized, "error extracting token")
		// Try to get the davs (must fail)
		do("GET", "/api/admin/davs", noH, "", http.StatusUnauthorized, "error extracting token")
		// Try to create a dav (must fail)
		do("POST", "/api/admin/davs/", noH, newDav, http.StatusUnauthorized, "error extracting token")
		// Try to delete a dav (must fail)
		do("DELETE", "/api/admin/davs/3", noH, "", http.StatusUnauthorized, "error extracting token")
		// Try to get the users (must fail)
		do("GET", "/api/admin/users/", noH, "", http.StatusUnauthorized, "error extracting token")
		// Try to create an user (must fail)
		do("POST", "/api/admin/users/", noH, newUser, http.StatusUnauthorized, "error extracting token")
		// Try to delete an user (must fail)
		do("DELETE", "/api/admin/users/0", noH, "", http.StatusUnauthorized, "error extracting token")
		// Try to access the apps list (must fail)
		do("GET", "/api/common/apps", noH, "", http.StatusUnauthorized, "error extracting token")
		// Try to access an app (must fail)
		do("GET", "api.vestibule.io", noH, "", http.StatusUnauthorized, "error extracting token")
		// Try to access the davs list (must fail)
		do("GET", "/api/common/davs", noH, "", http.StatusUnauthorized, "error extracting token")
		// Try to access a secured dav (must fail)
		do("GET", "userdav.vestibule.io/mydata/test.txt", noH, "", http.StatusUnauthorized, "error extracting token")
		// Try to access a non crypted file on an unsecured dav (must pass)
		do("GET", "unsecureddav.vestibule.io/mydata/test.txt", noH, "", http.StatusOK, "This is a test !")
		// Try to access a crypted file on a encrypted unsecured dav (must pass)
		do("PUT", "encrypteddav.vestibule.io/myencrypteddata/test-ciphered.txt", noH, "content is encrypted !", http.StatusCreated, "")
		// Try to access a crypted file on a encrypted unsecured dav (must pass)
		do("GET", "encrypteddav.vestibule.io/myencrypteddata/test-ciphered.txt", noH, "", http.StatusOK, "content is encrypted !")
		// Try to access a non crypted file on a encrypted unsecured dav (must fail)
		do("GET", "encrypteddav.vestibule.io/myencrypteddata/test.txt", noH, "", http.StatusInternalServerError, "unexpected EOF")
		// Try to access the main page (must pass)
		do("GET", "/", noH, "", http.StatusOK, "<!DOCTYPE html>")
		// Try to get the user informations (must fail)
		do("GET", "/api/common/WhoAmI", noH, "", http.StatusUnauthorized, "error extracting token")
		// Do a in memory login with an unknown user
		do("POST", "/Login", noH, `{"login": "unknownuser","password": "password"}`, http.StatusForbidden, `user not found`)
		// Do a in memory login with a known user but bad password
		do("POST", "/Login", noH, `{"login": "admin","password": "badpassword"}`, http.StatusForbidden, `user not found`)
		// Try to get a share token (must fail)
		do("POST", "/api/common/Share", noH, `{"sharedfor":"guest","url":"userdav.vestibule.io/mydata/test.txt","lifespan":1,"readonly":true}`, http.StatusUnauthorized, "error extracting token")
	}
}

/**
USER TESTS (this tests are to check that a normally logged user can access the apps that is allowed to and only that)
**/
func createUserTests(t *testing.T) func(wg *sync.WaitGroup) {
	// Create the tester
	ts, do, doNoJar := createTester(t)
	return func(wg *sync.WaitGroup) {
		defer ts.Close() // Close the tester
		defer wg.Done()
		tests := func() {
			// Get the XSRF Token
			response := do("GET", "/api/common/WhoAmI", noH, "", http.StatusOK, "")
			token := auth.TokenData{}
			json.Unmarshal([]byte(response), &token)
			xsrfHeader := map[string]string{"XSRF-TOKEN": token.XSRFToken}
			// Try to get the apps without XSRF token (must fail)
			do("GET", "/api/admin/apps", noH, "", http.StatusUnauthorized, "XSRF")
			// Try to get the apps (must fail)
			do("GET", "/api/admin/apps", xsrfHeader, "", http.StatusForbidden, "no user role among")
			// Try to create an app (must fail)
			do("POST", "/api/admin/apps/", xsrfHeader, newApp, http.StatusForbidden, "no user role among")
			// Try to delete an app (must fail)
			do("DELETE", "/api/admin/apps/4", xsrfHeader, "", http.StatusForbidden, "no user role among")
			// Try to get the davs as user (must pass, without passphrase)
			do("GET", "/api/common/davs", xsrfHeader, "", http.StatusOK, `[{"id":1,"host":"encrypteddav.vestibule.io","root":"./testdata/data/cipherdata","writable":true,"secured":false,"usedgb"`)
			// Try to get the davs as admin (must fail)
			do("GET", "/api/admin/davs", xsrfHeader, "", http.StatusForbidden, "no user role among")
			// Try to create a dav (must fail)
			do("POST", "/api/admin/davs/", xsrfHeader, newDav, http.StatusForbidden, "no user role among")
			// Try to delete a dav (must fail)
			do("DELETE", "/api/admin/davs/3", xsrfHeader, "", http.StatusForbidden, "no user role among")
			// Try to get the users (must fail)
			do("GET", "/api/admin/users/", xsrfHeader, "", http.StatusForbidden, "no user role among")
			// Try to create an user (must fail)
			do("POST", "/api/admin/users/", xsrfHeader, newUser, http.StatusForbidden, "no user role among")
			// Try to delete an user (must fail)
			do("DELETE", "/api/admin/users/0", xsrfHeader, "", http.StatusForbidden, "no user role among")
			// Try to access the apps list (must pass)
			do("GET", "/api/common/apps", xsrfHeader, "", http.StatusOK, "[{\"id\":1,\"isProxy\"")
			// Try to access a forbidden app (must fail)
			do("GET", "external.vestibule.io", xsrfHeader, "", http.StatusForbidden, "no user role among")
			// Try to access the davs list (must pass)
			do("GET", "/api/common/davs", xsrfHeader, "", http.StatusOK, "[{\"id\":1")
			// Try to access a forbidden dav (must fail)
			do("GET", "admindav.vestibule.io", xsrfHeader, "", http.StatusForbidden, "no user role among")
			// Try to access the main page (must pass)
			do("GET", "/", xsrfHeader, "", http.StatusOK, "<!DOCTYPE html>")
			// Try to access an authorized app (must pass)
			do("GET", "api.vestibule.io", xsrfHeader, "", http.StatusOK, "{")
			// Try to access an authorized dav (must pass)
			do("GET", "userdav.vestibule.io/mydata/test.txt?inline", xsrfHeader, "", http.StatusOK, "This is a test")
			// Try to create a ressource on the readonly dav (must fail)
			do("PUT", "userdav.vestibule.io/mydata/test2.txt", xsrfHeader, "This is a write test", http.StatusMethodNotAllowed, "method not allowed : dav is read only")
			// Try to get the user informations (must pass)
			do("GET", "/api/common/WhoAmI", xsrfHeader, "", http.StatusOK, `{"id":`)
			// Try to get a share token (must pass)
			shareHeader := map[string]string{"Authorization": "Bearer " + do("POST", "/api/common/Share", xsrfHeader, `{"sharedfor":"guest","url":"userdav.vestibule.io/mydata/test.txt","lifespan":1,"readonly":true}`, http.StatusOK, "")}
			// Try get the specified resource without cookie (must fail)
			doNoJar("GET", "userdav.vestibule.io/mydata/test.txt", xsrfHeader, "", http.StatusUnauthorized, "error extracting token")
			// Try to use the share token for the specified ressource (must pass)
			doNoJar("GET", "userdav.vestibule.io/mydata/test.txt", shareHeader, "", http.StatusOK, "This is a test")
			// Try to use the share token for the specified ressource with query (must pass)
			doNoJar("GET", "userdav.vestibule.io/mydata/test.txt?token="+url.QueryEscape(strings.TrimPrefix(shareHeader["Authorization"], "Bearer ")), noH, "", http.StatusOK, "This is a test")
			// Try to use the readonly share token to alter the specified ressource (must fail)
			doNoJar("PUT", "userdav.vestibule.io/mydata/test.txt", shareHeader, "Altered content", http.StatusForbidden, "token is read only")
			// Try to use the share token for an other ressource (must fail)*/
			doNoJar("GET", "userdav.vestibule.io/mydata/another_test.txt", shareHeader, "", http.StatusUnauthorized, "token restricted to url")
			// Try to get a share token for a forbidden -admin only- resource (should pass, but token should be useless)
			shareHeader = map[string]string{"Authorization": "Bearer " + do("POST", "/api/common/Share", xsrfHeader, `{"sharedfor":"guest","url":"admindav.vestibule.io/mydata/test.txt","lifespan":1,"readonly":true}`, http.StatusOK, "")}
			// Try to use the previous token for a forbidden resource (must fail)
			doNoJar("GET", "admindav.vestibule.io/mydata/test.txt", shareHeader, "", http.StatusForbidden, "no user role among")
		}
		// Try to login with OAuth2 (must pass)
		do("GET", "/OAuth2Login", noH, "", http.StatusOK, "<!DOCTYPE html>")
		// Run the tests
		tests()
		// Try to logout (must pass)
		do("GET", "/Logout", noH, "", http.StatusOK, "Logout OK")
		// Try to access an authorized app after logout (must fail)
		do("GET", "api.vestibule.io", noH, "", http.StatusUnauthorized, "error extracting token")
		// Do a in memory login with an known user
		do("POST", "/Login", noH, `{"login": "user","password": "password"}`, http.StatusOK, "")
		// Run the tests
		tests()
		// Try to logout (must pass)
		do("GET", "/Logout", noH, "", http.StatusOK, "Logout OK")
		// Try to access an authorized app after logout (must fail)
		do("GET", "api.vestibule.io", noH, "", http.StatusUnauthorized, "error extracting token")
	}
}

/**
ADMIN TESTS (this tests are to check that an administrator can alter the apps)
**/
func createAdminTests(t *testing.T) func(wg *sync.WaitGroup) {
	// Create the tester
	ts, do, _ := createTester(t)
	return func(wg *sync.WaitGroup) {
		defer ts.Close() // Close the tester
		defer wg.Done()
		tests := func() {
			// Get the XSRF Token
			response := do("GET", "/api/common/WhoAmI", noH, "", http.StatusOK, "")
			token := auth.TokenData{}
			json.Unmarshal([]byte(response), &token)
			xsrfHeader := map[string]string{"XSRF-TOKEN": token.XSRFToken}
			// Try to get the apps (must pass)
			do("GET", "/api/admin/apps/", xsrfHeader, "", http.StatusOK, "[{\"id\":1")
			// Try to create an app without the XSRF-TOKEN (must fail)
			do("POST", "/api/admin/apps/", noH, newApp, http.StatusUnauthorized, "XSRF")
			// Try to create an app (must pass)
			do("POST", "/api/admin/apps/", xsrfHeader, newApp, http.StatusOK, "[{\"id\":1")
			// Try to delete an app (must pass)
			do("DELETE", "/api/admin/apps/4", xsrfHeader, "", http.StatusOK, "[{\"id\":1")
			// Try to get the davs (must pass, with passphrase)
			do("GET", "/api/admin/davs/", xsrfHeader, "", http.StatusOK, `[{"id":1,"host":"encrypteddav.vestibule.io","root":"./testdata/data/cipherdata","writable":true,"secured":false,"passphrase":"very secret passphrase"`)
			// Try to create a dav (must pass)
			do("POST", "/api/admin/davs/", xsrfHeader, newDav, http.StatusOK, "[{\"id\":1")
			// Try to delete a dav (must pass)
			do("DELETE", "/api/admin/davs/5", xsrfHeader, "", http.StatusOK, "[{\"id\":1")
			// Try to get the users (must pass)
			do("GET", "/api/admin/users/", xsrfHeader, "", http.StatusOK, `[{"id":"1",`)
			// Try to create an user (must pass)
			do("POST", "/api/admin/users/", xsrfHeader, newUser, http.StatusOK, `[{"id":"1",`)
			// Try to recreate the same user (must pass, it will update)
			do("POST", "/api/admin/users/", xsrfHeader, newUser, http.StatusOK, `[{"id":"1",`)
			// Try to create an user with an existing login (must fail)
			do("POST", "/api/admin/users/", xsrfHeader, `{"id":"4","login":"new_user","memberOf":["USERS"],"password":"test"}`, http.StatusBadRequest, `login already exists`)
			// Try to delete an user (must pass)
			do("DELETE", "/api/admin/users/3", xsrfHeader, "", http.StatusOK, `[{"id":"1",`)
			// Try to log with the deleted user (must fail)
			do("POST", "/Login", noH, `{"login": "new_user","password": "test"}`, http.StatusForbidden, "")
			// Try to get the user informations (must pass)
			do("GET", "/api/common/WhoAmI", xsrfHeader, "", http.StatusOK, `{"id":`)
			// Try to access an authorized dav (must pass)
			do("GET", "admindav.vestibule.io/mydata/test.txt", xsrfHeader, "", http.StatusOK, "This is a test")
			// Try to alter a resource on an authorized dav without XSRF Token (must fail)
			do("PUT", "admindav.vestibule.io/mydata/test2.txt", noH, "This is a write test", http.StatusUnauthorized, "XSRF")
			// Try to alter a resource on an authorized dav (must pass)
			do("PUT", "admindav.vestibule.io/mydata/test2.txt", xsrfHeader, "This is a write test", http.StatusCreated, "")
			// Try to delete a resource on an authorized dav (must pass)
			do("DELETE", "admindav.vestibule.io/mydata/test2.txt", xsrfHeader, "", http.StatusNoContent, "")
			// Try to get the system information (must pass)
			if runtime.GOOS == "windows" {
				do("GET", "/api/admin/sysinfo/", xsrfHeader, "", http.StatusOK, `{"usedgb"`)
			} else {
				do("GET", "/api/admin/sysinfo/", xsrfHeader, "", http.StatusOK, `{"uptime"`)
			}
		}
		// Try to login (must pass)
		do("GET", "/OAuth2Login", noH, "", http.StatusOK, "<!DOCTYPE html>")
		// Run the tests
		tests()
		// Try to logout (must pass)
		do("GET", "/Logout", noH, "", http.StatusOK, "Logout OK")
		// Try to get the apps again (must fail)
		do("GET", "/api/admin/apps", noH, "", http.StatusUnauthorized, "error extracting token")
		// Do a in memory login with an known admin
		do("POST", "/Login", noH, `{"login": "admin","password": "password"}`, http.StatusOK, "")
		tests()
		// Try to logout (must pass)
		do("GET", "/Logout", noH, "", http.StatusOK, "Logout OK")
		// Try to get the apps again (must fail)
		do("GET", "/api/admin/apps", noH, "", http.StatusUnauthorized, "error extracting token")
	}
}

/**
DIRECT WEBDAV TESTS (this tests are to check that a direct access webdav works)
**/
func createDirectWebdavTests(t *testing.T) func(wg *sync.WaitGroup) {
	// Users
	authFromUser := func(user auth.User) string {
		data := user.Login + ":" + user.Password
		return "Basic " + b64.StdEncoding.EncodeToString([]byte(data))
	}
	correctAuthHeader := map[string]string{"Authorization": authFromUser(auth.User{Login: "user", Password: "password"}), "User-Agent": "LibreOffice"}
	wrongAuthHeader := map[string]string{"Authorization": authFromUser(auth.User{Login: "user", Password: "wrong_password"}), "User-Agent": "LibreOffice"}
	wrongUserAgent := map[string]string{"Authorization": authFromUser(auth.User{Login: "user", Password: "password"}), "User-Agent": "Other"}
	// Create the tester
	ts, do, _ := createTester(t)
	return func(wg *sync.WaitGroup) {
		defer ts.Close() // Close the tester
		defer wg.Done()
		// Try to get the apps (must fail)
		do("GET", "/api/admin/apps", correctAuthHeader, "", http.StatusForbidden, "no user role among")
		// Try to access the davs list (must pass)
		do("GET", "/api/common/davs", correctAuthHeader, "", http.StatusOK, "[{\"id\":1")
		// Try to access a forbidden dav (must fail)
		do("GET", "admindav.vestibule.io", correctAuthHeader, "", http.StatusForbidden, "no user role among")
		// Try to access the main page (must pass)
		do("GET", "/", correctAuthHeader, "", http.StatusOK, "<!DOCTYPE html>")
		// Try to access an authorized app (must pass)
		do("GET", "api.vestibule.io", correctAuthHeader, "", http.StatusOK, "{")
		// Try to access an authorized dav (must pass)
		do("GET", "userdav.vestibule.io/mydata/test.txt?inline", correctAuthHeader, "", http.StatusOK, "This is a test")
		// Try to access a forbidden dav  (must fail)
		do("GET", "admindav.vestibule.io/mydata/test.txt", correctAuthHeader, "", http.StatusForbidden, "no user role among")
		// Try to access an authorized dav with wrong password (must fail)
		do("GET", "userdav.vestibule.io/mydata/test.txt?inline", wrongAuthHeader, "", http.StatusUnauthorized, "webdav client authentication")
		// Try to access an authorized dav with wrong user agent (must fail)
		do("GET", "userdav.vestibule.io/mydata/test.txt?inline", wrongUserAgent, "", http.StatusUnauthorized, "error extracting token")

	}
}

func createTester(t *testing.T) (*httptest.Server, tester.DoFn, tester.DoFn) {
	// Create the server
	mux := CreateRootMux(os.Getenv("HOSTNAME"), 1443, "./testdata/apps.json", "./testdata/davs.json", "../../web")
	ts := httptest.NewServer(mux.Mux)
	url, _ := url.Parse(ts.URL)
	port := url.Port()
	mux.Manager.Config.RedirectURL = "http://" + os.Getenv("HOSTNAME") + ":" + port + "/OAuth2Callback"
	mux.Manager.Hostname = "http://" + os.Getenv("HOSTNAME") + ":" + port
	// Create the cookie jar
	jar, _ := cookiejar.New(nil)
	// wrap the testing function
	return ts, tester.CreateServerTester(t, port, os.Getenv("HOSTNAME"), jar, true), tester.CreateServerTester(t, port, os.Getenv("HOSTNAME"), nil, true)
}
