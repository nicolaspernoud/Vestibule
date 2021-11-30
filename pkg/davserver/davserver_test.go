package davserver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nicolaspernoud/vestibule/pkg/tester"
)

func TestEncryption(t *testing.T) {
	// Create the handler
	davAug := NewWebDavAug("", "./testdata", true, "very secret passphrase")
	ts := httptest.NewServer(&davAug)
	url, _ := url.Parse(ts.URL)
	port := url.Port()
	// wrap the testing function
	do := tester.CreateServerTester(t, port, "vestibule.io", nil, true)
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

func TestSetModTime(t *testing.T) {
	test := func(do tester.DoFn) {
		// Time is now
		now := time.Now()
		time.Sleep(1 * time.Second)
		// No X-OC-Mtime header
		var noH map[string]string
		// Try to put a file without the Mtime header (must pass)
		do("PUT", "/test-modtime.txt", noH, "no Mtime", 201, "")
		// Check that the modified time of the file is greater than now
		fi, _ := os.Stat("./testdata/test-modtime.txt")
		mtime := fi.ModTime()
		if mtime.Before(now) {
			t.Errorf("modification time must be greater than reference time")
		}
		// Try to put a file with a Mtime header ()
		mtH := map[string]string{
			"X-OC-Mtime": "405659700",
		}
		do("PUT", "/test-modtime.txt", mtH, "Mtime is my birthday", 201, "")
		// Check that the modifier time of the file matches the header
		fi, _ = os.Stat("./testdata/test-modtime.txt")
		mtime = fi.ModTime()
		myBD := time.Unix(405659700, 0)
		if !mtime.Equal(myBD) {
			t.Errorf("modification time must be lesser than reference time")
		}
	}

	// Test with unencrypted webdav
	davAug := NewWebDavAug("", "./testdata", true, "")
	ts := httptest.NewServer(&davAug)
	url, _ := url.Parse(ts.URL)
	port := url.Port()
	doPlain := tester.CreateServerTester(t, port, "vestibule.io", nil, true)
	test(doPlain)

	// Test with encrypted webdav
	davAugEnc := NewWebDavAug("", "./testdata", true, "this is encrypted")
	tsEnc := httptest.NewServer(&davAugEnc)
	urlEnc, _ := url.Parse(tsEnc.URL)
	portEnc := urlEnc.Port()
	doEnc := tester.CreateServerTester(t, portEnc, "vestibule.io", nil, true)
	test(doEnc)
}
