package tester

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type DoFn func(method string, url string, headers map[string]string, payload string, expectedStatus int, expectedBody string) string

// DoRequestOnHandler does a request on a router (or handler) and check the response
func DoRequestOnHandler(t *testing.T, router http.Handler, method string, route string, headers map[string]string, payload string, expectedStatus int, expectedBody string) string {
	req, err := http.NewRequest(method, route, strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range headers {
		req.Header.Set(i, v)
	}
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if status := rr.Code; status != expectedStatus {
		t.Errorf("Tested %v %v %v ; handler returned wrong status code: got %v want %v", method, route, payload, status, expectedStatus)
	}
	if !strings.HasPrefix(rr.Body.String(), expectedBody) {
		t.Errorf("Tested %v %v %v ; handler returned unexpected body: got %v want %v", method, route, payload, rr.Body.String(), expectedBody)
	}
	return string(rr.Body.String())
}

// DoRequestOnServer does a request on listening server
func DoRequestOnServer(t *testing.T, hostname string, port string, jar *cookiejar.Jar, method string, testURL string, headers map[string]string, payload string, expectedStatus int, expectedBody string) string {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	// or create your own transport, there's an example on godoc.
	http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		addrAndPort := strings.Split(addr, ":")
		if strings.HasSuffix(addrAndPort[0], "vestibule.io") {
			addr = "127.0.0.1:" + addrAndPort[1]
		}
		return dialer.DialContext(ctx, network, addr)
	}
	if strings.HasPrefix(testURL, "/") {
		testURL = "http://" + hostname + ":" + port + testURL
	} else {
		u, _ := url.Parse("http://" + testURL)
		testURL = "http://" + u.Host + ":" + port + u.Path + "?" + u.RawQuery
	}
	req, err := http.NewRequest(method, testURL, strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range headers {
		req.Header.Set(i, v)
	}
	var client *http.Client
	if jar != nil {
		client = &http.Client{Jar: jar}
	} else {
		client = &http.Client{}
	}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := ioutil.ReadAll(res.Body)
	bodyString := string(body)
	if status := res.StatusCode; status != expectedStatus {
		t.Errorf("Tested %v %v %v ; handler returned wrong status code: got %v want %v", method, testURL, payload, status, expectedStatus)
	}
	if !strings.HasPrefix(bodyString, expectedBody) {
		t.Errorf("Tested %v %v %v ; handler returned unexpected body: got %v want %v", method, testURL, payload, bodyString, expectedBody)
	}
	return bodyString
}

// CreateServerTester wraps DoRequestOnServer to factorize t, port and jar
func CreateServerTester(t *testing.T, hostname string, port string, jar *cookiejar.Jar) DoFn {
	return func(method string, url string, headers map[string]string, payload string, expectedStatus int, expectedBody string) string {
		return DoRequestOnServer(t, port, hostname, jar, method, url, headers, payload, expectedStatus, expectedBody)
	}
}
