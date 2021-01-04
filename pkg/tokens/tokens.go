package tokens

import (
	"bytes"
	"compress/flate"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/nicolaspernoud/vestibule/pkg/common"
	"github.com/nicolaspernoud/vestibule/pkg/log"
)

var (
	now = time.Now
	// Manager is the current token manager
	Manager manager
)

// manager manages tokens
type manager struct {
	key       []byte
	debugMode bool
}

// Init inits the main token manager
func Init(keyfile string, debug bool) {
	Manager = newManager(keyfile, debug)
}

// newManager creates a manager
func newManager(keyfile string, debug bool) manager {
	var keyConfig struct {
		Key []byte
	}
	err := common.Load(keyfile, &keyConfig)
	if err != nil {
		keyConfig.Key, err = common.GenerateRandomBytes(32)
		if err != nil {
			log.Logger.Fatal(err)
		}
		err := common.Save(keyfile, keyConfig)
		if err != nil {
			log.Logger.Println("Token signing key could not be saved")
		}
	}
	log.Logger.Println("Token signing key set")
	return manager{
		debugMode: debug,
		key:       keyConfig.Key,
	}
}

// Token represents a token containting data
type Token struct {
	ExpiresAt int64
	IssuedAt  int64 `json:"iat,omitempty"`
	Data      []byte
}

// StoreData creates a token with the given data and returns it in a cookie
func (m manager) StoreData(data interface{}, hostName string, cookieName string, duration time.Duration, w http.ResponseWriter) {
	expiration := now().Add(duration)
	value, err := m.CreateToken(data, expiration)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	cookie := http.Cookie{Name: cookieName, Domain: hostName, Value: value, Expires: expiration, Secure: !m.debugMode, HttpOnly: true, SameSite: http.SameSiteLaxMode}
	http.SetCookie(w, &cookie)
}

// CreateToken creates a token with the given data
func (m manager) CreateToken(data interface{}, expiration time.Time) (string, error) {
	// Marshall the data
	d, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	// Create the payload
	token := Token{
		ExpiresAt: expiration.Unix(),
		Data:      d,
	}
	// Serialize the payload
	sToken, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	// Compress with deflate
	var csToken bytes.Buffer
	c, err := flate.NewWriter(&csToken, flate.BestCompression)
	if _, err := c.Write(sToken); err != nil {
		return "", err
	}
	if err := c.Close(); err != nil {
		return "", err
	}
	ecsToken, err := Encrypt(csToken.Bytes(), m.key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ecsToken), nil
}

// ExtractAndValidateToken extracts the token from the request, validates it, and return the data n the value pointed to by v
func (m manager) ExtractAndValidateToken(r *http.Request, cookieName string, v interface{}, checkXSRF bool) (bool, error) {
	becsToken, checkXSRF, err := func(r *http.Request, checkXSRF bool) (string, bool, error) {
		// Try to extract from the query
		query := r.URL.Query().Get("token")
		if query != "" {
			return query, false, nil
		}
		// Try to extract from the cookie
		cookie, err := r.Cookie(cookieName)
		if err == nil {
			return cookie.Value, checkXSRF, err
		}
		// Try to get an auth token from the header
		authHeader := strings.Split(r.Header.Get("Authorization"), " ")
		if authHeader[0] == "Bearer" && len(authHeader) == 2 {
			return authHeader[1], false, nil
		}
		// Try to use the basic auth header instead
		if authHeader[0] == "Basic" && len(authHeader) == 2 {
			decoded, err := base64.StdEncoding.DecodeString(authHeader[1])
			if err == nil {
				auth := strings.Split(string(decoded), ":")
				return auth[1], false, nil
			}
		}
		return "", false, errors.New("could not extract token")
	}(r, checkXSRF)

	if err == nil {
		return checkXSRF, m.unstoreData(becsToken, v)
	}
	return false, err
}

// unstoreData decrypt, uncompress, unserialize the token, and returns the data n the value pointed to by v
func (m manager) unstoreData(becsToken string, v interface{}) error {
	// Decrypt the token
	ecsToken, err := base64.StdEncoding.DecodeString(becsToken)
	if err != nil {
		return fmt.Errorf("failed to unbase64 token")

	}
	csToken, err := Decrypt(ecsToken, m.key)
	if err != nil {
		return fmt.Errorf("failed to decrypt token")

	}
	// Uncompress the token
	rdata := bytes.NewReader(csToken)
	r := flate.NewReader(rdata)
	sToken, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to uncompress token")

	}
	// Unserialize the token
	token := Token{}
	err = json.Unmarshal(sToken, &token)
	if err != nil {
		return fmt.Errorf("failed to unmarshall token")

	}
	// Validate the token
	if token.ExpiresAt < now().Unix() {
		return fmt.Errorf("token expired")
	}
	// Update the data
	err = json.Unmarshal(token.Data, v)
	// Return no error if everything is fine
	return nil
}

// Encrypt a byte array with AES
func Encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return []byte{}, err
	}
	cipherData := gcm.Seal(nonce, nonce, data, nil)
	return cipherData, nil
}

// Decrypt a byte array with AES
func Decrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) <= nonceSize {
		return []byte{}, err
	}
	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plainData, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return []byte{}, err
	}
	return plainData, nil
}
