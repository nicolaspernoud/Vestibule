package davserver

import (
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/nicolaspernoud/vestibule/pkg/tester"
)

func TestEncryption(t *testing.T) {
	// Create the handler
	davAug := NewWebDavAug("", "./testdata", true, "very secret passphrase")
	ts := httptest.NewServer(&davAug)
	url, _ := url.Parse(ts.URL)
	port := url.Port()
	// wrap the testing function
	do := tester.CreateServerTester(t, port, os.Getenv("HOSTNAME"), nil)
	noH := tester.Header{Key: "", Value: ""}
	// Try to access a crypted file on a encrypted unsecured dav (must pass)
	do("PUT", "/test-ciphered.txt", noH, "content is encrypted !", 201, "")
	// Try to access a crypted file on a encrypted unsecured dav (must pass)
	do("GET", "/test-ciphered.txt", noH, "", 200, "content is encrypted !")
	// Try to access a non crypted file on a encrypted unsecured dav (must fail)
	do("GET", "/test.txt", noH, "", 500, "unexpected EOF")
	// Try to access a crypted file with the wrong key
	davAug = NewWebDavAug("", "./testdata", true, "wrong key")
	body := do("GET", "/test-ciphered.txt", noH, "", 500, "")
	if body != "" { // Check that the body is really empty
		t.Errorf("body must be empty")
	}
}
