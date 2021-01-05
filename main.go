package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nicolaspernoud/vestibule/pkg/common"
	"github.com/nicolaspernoud/vestibule/pkg/middlewares"
	"github.com/nicolaspernoud/vestibule/pkg/tokens"

	"github.com/nicolaspernoud/vestibule/internal/mocks"
	"github.com/nicolaspernoud/vestibule/internal/rootmux"
	"github.com/nicolaspernoud/vestibule/pkg/log"

	"golang.org/x/crypto/acme/autocert"
)

var (
	appsFile, davsFile, letsCacheDir, logFile string
	httpsPort, httpPort                       int
	debugMode                                 bool
)

func init() {
	var err error
	appsFile, err = common.StringValueFromEnv("APPS_FILE", "./configs/apps.json") // Apps configuration file path
	common.CheckErrorFatal(err)
	davsFile, err = common.StringValueFromEnv("DAVS_FILE", "./configs/davs.json") // Davs configuration file path
	common.CheckErrorFatal(err)
	letsCacheDir, err = common.StringValueFromEnv("LETS_CACHE_DIR", "./letsencrypt_cache") // Let's Encrypt cache directory
	common.CheckErrorFatal(err)
	logFile, err = common.StringValueFromEnv("LOG_FILE", "") // Optional file to log to
	common.CheckErrorFatal(err)
	httpsPort, err = common.IntValueFromEnv("HTTPS_PORT", 443) // HTTPS port to serve on
	common.CheckErrorFatal(err)
	httpPort, err = common.IntValueFromEnv("HTTP_PORT", 80) // HTTP port to serve on, only used for Let's Encrypt HTTP Challenge
	common.CheckErrorFatal(err)
	debugMode, err = common.BoolValueFromEnv("DEBUG_MODE", false) // Debug mode, disable Let's Encrypt, enable CORS and more logging
	common.CheckErrorFatal(err)
}

func main() {
	// Initialize logger
	if logFile != "" {
		log.SetFile(logFile)
		// Properly close the log on exit
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigs
			log.Logger.Println("--- Closing log ---")
			log.CloseFile()
			os.Exit(0)
		}()
	}
	log.Logger.Println("--- Server is starting ---")
	fullHostname := middlewares.GetFullHostname(os.Getenv("HOSTNAME"), httpsPort)
	log.Logger.Println("Main hostname is ", fullHostname)

	// Initializations
	tokens.Init("./configs/tokenskey.json", debugMode)

	// Create the server
	rootMux := rootmux.CreateRootMux(httpsPort, appsFile, davsFile, "web")

	// Serve locally with https on debug mode or with let's encrypt on production mode
	if debugMode {
		// Init the hostname
		mocks.Init(httpsPort)
		// Start a mock oauth2 server if debug mode is on
		mockOAuth2Port := ":8090"
		go http.ListenAndServe(mockOAuth2Port, mocks.CreateMockOAuth2())
		fmt.Println("Mock OAuth2 server Listening on: http://localhost" + mockOAuth2Port)
		// Start a mock API server if debug mode is on
		mockAPIPort := ":8091"
		go http.ListenAndServe(mockAPIPort, mocks.CreateMockAPI())
		fmt.Println("Mock API server Listening on: http://localhost" + mockAPIPort)
		log.Logger.Fatal(http.ListenAndServeTLS(":"+strconv.Itoa(httpsPort), "./dev_certificates/localhost.crt", "./dev_certificates/localhost.key", log.Middleware(rootMux.Mux)))

	} else {
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache(letsCacheDir),
			HostPolicy: rootMux.Policy,
		}

		server := &http.Server{
			Addr:    ":" + strconv.Itoa(httpsPort),
			Handler: rootMux.Mux,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
				MinVersion:     tls.VersionTLS12,
			},
			ReadTimeout:  30 * time.Minute, // in case of upload
			WriteTimeout: 5 * time.Hour,    // in case of download
			IdleTimeout:  120 * time.Second,
		}

		go http.ListenAndServe(":"+strconv.Itoa(httpPort), certManager.HTTPHandler(nil))
		log.Logger.Fatal(server.ListenAndServeTLS("", ""))
	}
}
