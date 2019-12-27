package jwt

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/nicolaspernoud/vestibule/pkg/common"
	"github.com/nicolaspernoud/vestibule/pkg/log"
)

type key int

const (
	// ContextData is the data got from the JWT
	ContextData key = 0
)

var (
	jWTSignature []byte
	now          = time.Now
	debugMode    bool
)

// AuthToken represents a token identifying an data
type AuthToken struct {
	jwt.StandardClaims
	Data      interface{}
	CSRFToken string `json:"csrftoken"`
}

// Init sets the jWTSignature
func Init(debug bool) {
	debugMode = debug
	var jWTConfig struct {
		JWTSignature string
	}
	err := common.Load("./configs/jwtsignature.json", &jWTConfig)
	if err != nil {
		jWTConfig.JWTSignature, err = common.GenerateRandomString(48)
		if err != nil {
			log.Logger.Fatal(err)
		}
		err := common.Save("./configs/jwtsignature.json", jWTConfig)
		if err != nil {
			log.Logger.Println("Token signing key could not be saved")
		}
	}
	jWTSignature = []byte(jWTSignature)
	log.Logger.Println("Token signing key set")
}

// StoreData stores the given data in a jwt cookie
func StoreData(d interface{}, hostName string, cookieName string, duration time.Duration, w http.ResponseWriter) {
	expiration := now().Add(duration)
	CSRFToken, err := common.GenerateRandomString(16)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, AuthToken{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiration.Unix(),
			IssuedAt:  now().Unix(),
		},
		Data:      d,
		CSRFToken: CSRFToken,
	})
	tokenString, err := token.SignedString(jWTSignature)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Compress the token with gzip
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(tokenString)); err != nil {
		http.Error(w, "failed to compress token", 500)
		return
	}
	if err := gz.Close(); err != nil {
		http.Error(w, "failed to compress token", 500)
		return
	}
	cookie := http.Cookie{Name: cookieName, Domain: hostName, Value: base64.StdEncoding.EncodeToString(b.Bytes()), Expires: expiration, Secure: !debugMode, HttpOnly: true, SameSite: http.SameSiteLaxMode}
	http.SetCookie(w, &cookie)
	//log.Logger.Printf("| %v (%v %v) | Login success | %v | %v", sentData.Login, data.Name, data.Surname, req.RemoteAddr, log.GetCityAndCountryFromRequest(req))
}

// ExtractToken extracts the jwt from a cookie
func ExtractToken(r *http.Request, cookieName string) (string, error) {
	// Try to get an auth token from the cookie
	jwtCookie, err := r.Cookie(cookieName)
	if err == nil {
		data, err2 := base64.StdEncoding.DecodeString(jwtCookie.Value)
		if err2 != nil {
			return "", fmt.Errorf("failed to uncompress token")
		}
		rdata := bytes.NewReader(data)
		r, _ := gzip.NewReader(rdata)
		s, _ := ioutil.ReadAll(r)
		return string(s), nil
	}
	return "", fmt.Errorf("no token found")
}

// ValidateJWTMiddleware tests if a JWT token is present and valid in the request
// It returns the data has request context
func ValidateJWTMiddleware(next http.Handler, cookieName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		JWT, err := ExtractToken(req, cookieName)
		if err != nil {
			http.Error(w, fmt.Sprintf("error extracting JWT: %v", err), 401)
			return
		}
		token, err := jwt.ParseWithClaims(JWT, &AuthToken{}, checkJWT)
		if err != nil {
			http.Error(w, "error parsing JWT", 401)
			return
		}
		claims, ok := token.Claims.(*AuthToken)
		if ok && token.Valid {
			// check for CSRF protection
			/*********
			**********
			IMPORTANT SECURITY ISSUE : test X-XSRF-TOKEN
			***********
			***********/
			/*if claims.CSRFToken != req.Header.Get("X-XSRF-TOKEN") {
				http.Error(w, "XSRF protection triggered", 403)
				return
			}*/
			ctx := context.WithValue(req.Context(), ContextData, claims.Data)
			next.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		http.Error(w, "invalid authorization token", 400)
	})
}

func checkJWT(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return jWTSignature, nil
}
