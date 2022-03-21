package common

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/nicolaspernoud/vestibule/pkg/log"
)

var (
	disableLogFatal = false
	lock            sync.Mutex // Mutex used to lock file writing
)

// Save saves a representation of v to the file at path.
func Save(path string, v interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r, err := Marshal(v)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}

// Load loads the file at path into v. Use os.IsNotExist() to see if the returned error is due to the file being missing.
func Load(path string, v interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return Unmarshal(f, v)
}

// Marshal is a function that marshals the object into an io.Reader. By default, it uses the JSON marshaller.
var Marshal = func(v interface{}) (io.Reader, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

// Unmarshal is a function that unmarshals the data from the reader into the specified value. By default, it uses the JSON unmarshaller.
var Unmarshal = func(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

// FallBackWrapper serves a file if found and else default to index.html
type FallBackWrapper struct {
	Assets http.FileSystem
}

// Open serves a file if found and else default to index.html
func (i *FallBackWrapper) Open(name string) (http.File, error) {
	file, err := i.Assets.Open(name)
	// If the file is found but there is another error or the asked for file has an extension : return the file or error
	if !os.IsNotExist(err) || path.Ext(name) != "" {
		return file, err
	}
	// Else fall back to index.html
	return i.Assets.Open("index.html")
}

// Contains works out if a slice contains a given element
func Contains[K comparable](a []K, x K) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// StringValueFromEnv set a value into an *interface from an environment variable or default
func StringValueFromEnv(ev string, def string) string {
	val := os.Getenv(ev)
	if val == "" {
		return def
	}
	return val
}

// IntValueFromEnv set a value into an *interface from an environment variable or default
func IntValueFromEnv(ev string, def int) int {
	val := os.Getenv(ev)
	if val == "" {
		return def
	}
	v, err := strconv.Atoi(val)
	if err != nil && !disableLogFatal {
		log.Logger.Fatalf("Error : could not get integer value from environment variable %v=%v\n", ev, val)
	}
	return v
}

// BoolValueFromEnv set a value into an *interface from an environment variable or default
func BoolValueFromEnv(ev string, def bool) bool {
	val := os.Getenv(ev)
	if val == "" {
		return def
	}
	v, err := strconv.ParseBool(val)
	if err != nil && !disableLogFatal {
		log.Logger.Fatalf("Error : could not get boolean value from environment variable %v=%v\n", ev, val)
	}
	return v
}
