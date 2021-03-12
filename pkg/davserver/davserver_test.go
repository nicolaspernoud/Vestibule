package davserver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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
	do := tester.CreateServerTester(t, port, "vestibule.io", nil)
	var noH map[string]string
	// Try to access a crypted file on a encrypted unsecured dav (must pass)
	do("PUT", "/test-ciphered.txt", noH, "content is encrypted !", 201, "")
	// Try to access a crypted file on a encrypted unsecured dav (must pass)
	do("GET", "/test-ciphered.txt", noH, "", 200, "content is encrypted !")
	// Try to get the true (unencrypted) file size on a encrypted unsecured dav (must pass)
	body := do("PROPFIND", "/test-ciphered.txt", noH, "", 207, "")
	if !strings.Contains(body, "<D:getcontentlength>22</D:getcontentlength>") {
		t.Errorf("test-ciphered.txt should be 22 bytes")
	}
	// Try to access a non crypted file on a encrypted unsecured dav (must fail)
	do("GET", "/test.txt", noH, "", http.StatusInternalServerError, "unexpected EOF")
	// Try to access a crypted file with the wrong key
	davAug = NewWebDavAug("", "./testdata", true, "wrong key")
	body = do("GET", "/test-ciphered.txt", noH, "", 200, "")
	if body != "" { // Check that the body is really empty
		t.Errorf("body must be empty")
	}
}
