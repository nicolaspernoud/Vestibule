package appserver

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/nicolaspernoud/vestibule/pkg/tester"
)

func TestServer(t *testing.T) {
	// Create the proxy target servers
	target := httptest.NewServer(http.HandlerFunc(testHandler))
	defer target.Close()

	// For the redirectFwdToTarget, we need to know the port in advance, so we use a custom listener
	// create a listener with the desired port.
	l, err := net.Listen("tcp", "127.0.0.1:8044")
	if err != nil {
		log.Fatal(err)
	}
	redirectFwdToTarget := httptest.NewUnstartedServer(http.HandlerFunc(testFwdToRedirectHandler))
	defer redirectFwdToTarget.Close()
	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	redirectFwdToTarget.Listener.Close()
	redirectFwdToTarget.Listener = l
	// Start the server.
	redirectFwdToTarget.Start()

	// Create the other servers (ports can be random)
	redirectRelativeTarget := httptest.NewServer(http.HandlerFunc(testRelativeRedirectHandler))
	defer redirectRelativeTarget.Close()

	redirectAbsoluteTarget := httptest.NewServer(http.HandlerFunc(testAbsoluteRedirectHandler))
	defer redirectAbsoluteTarget.Close()

	// Create apps
	appFile := writeApps([]*app{
		{App: App{Host: "test.proxy", IsProxy: true, ForwardTo: target.Listener.Addr().String()}},
		{App: App{Host: "*.test.wildcard", IsProxy: true, ForwardTo: target.Listener.Addr().String()}},
		{App: App{Host: "test.static", IsProxy: false, Serve: "testdata"}},
		{App: App{Host: "test.fwdtoredirect", IsProxy: true, ForwardTo: "127.0.0.1:8044"}},
		{App: App{Host: "test.relativeredirect", IsProxy: true, ForwardTo: redirectRelativeTarget.Listener.Addr().String()}},
		{App: App{Host: "test.absoluteredirect", IsProxy: true, ForwardTo: redirectAbsoluteTarget.Listener.Addr().String()}},
	})
	defer os.Remove(appFile)

	s, err := NewServer(appFile, 443, "localhost", "localhost", mockAuth)
	if err != nil {
		t.Fatal(err)
	}

	// Create tests
	var tests = []struct {
		url        string
		authHeader string
		code       int
		body       string
	}{
		{"http://test.proxy/", "", 200, "OK"},
		{"http://foo.test.proxy/", "", 404, "Not found."},
		{"http://footest.proxy/", "", 404, "Not found."},
		{"http://test.wildcard/", "", 200, "OK"},
		{"http://foo.test.wildcard/", "", 200, "OK"},
		{"http://test.static/", "", 200, "contents of index.html"},
		{"http://test.net/", "", 404, "Not found."},
	}

	// Run tests
	for _, test := range tests {
		tester.DoRequestOnHandler(t, s, "GET", test.url, test.authHeader, "", test.code, test.body)
	}

	// Create redirect tests
	var redirectTests = []struct {
		url      string
		code     int
		location string
	}{
		{"http://test.fwdtoredirect", 302, "https://test.fwdtoredirect:443/some/path"},
		{"http://test.relativeredirect/", 302, "https://relative.redirect.test.relativeredirect"},
		{"http://test.absoluteredirect/", 302, "https://absolute.redirect"},
	}

	// Run redirect tests
	for _, test := range redirectTests {
		rw := httptest.NewRecorder()
		rw.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", test.url, nil)
		s.ServeHTTP(rw, req)
		if g, w := rw.Code, test.code; g != w {
			t.Errorf("%s: code = %d, want %d", test.url, g, w)
		}
		if g, w := rw.Header().Get("Location"), test.location; g != w {
			t.Errorf("%s: location header = %q, want %q", test.url, g, w)
		}
	}
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// Redirect is bad when is made to the proxied host (fwdTo) and not to the exposed host (fwdFrom)
func testFwdToRedirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://fwdto.redirect.bad.127.0.0.1:8044/some/path", http.StatusFound)
}

// Redirect is good when is made to the host (fwdFrom)
func testRelativeRedirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://relative.redirect."+r.Host, http.StatusFound)
}

// Redirect is also good when is absolute (no links to neither the host -fwdFrom- or the proxied service -fwdTo-)
func testAbsoluteRedirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://absolute.redirect", http.StatusFound)
}

func writeApps(apps []*app) (name string) {
	f, err := ioutil.TempFile("", "vestibule-apps")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(apps)
	if err != nil {
		panic(err)
	}
	return f.Name()
}

func mockAuth(next http.Handler, allowedRoles []string) http.Handler {
	return next
}
