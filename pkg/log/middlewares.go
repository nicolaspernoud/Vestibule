package log

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// Middleware allow extensive logging of requests for debug and development purposes only
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		readBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			Logger.Print("Body error : ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newBody := ioutil.NopCloser(bytes.NewBuffer(readBody))
		r.Body = newBody
		Logger.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
		if string(readBody) != "" {
			Logger.Printf("BODY : %q", readBody)
		}
		next.ServeHTTP(w, r)
	})
}
