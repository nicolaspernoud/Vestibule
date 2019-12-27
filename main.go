package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nicolaspernoud/vestibule/pkg/jwt"

	"github.com/nicolaspernoud/vestibule/internal/mocks"
	"github.com/nicolaspernoud/vestibule/internal/rootmux"
	"github.com/nicolaspernoud/vestibule/pkg/log"

	"golang.org/x/crypto/acme/autocert"
)

var (
	appsFile     = flag.String("apps", "", "apps definition `file`")
	letsCacheDir = flag.String("letsencrypt_cache", "./letsencrypt_cache", "let's encrypt cache `directory`")
	logFile      = flag.String("log_file", "", "Optional file to log to, defaults to no file logging")
	httpsPort    = flag.Int("https_port", 443, "HTTPS port to serve on (defaults to 443)")
	httpPort     = flag.Int("http_port", 80, "HTTP port to serve on (defaults to 80), only used to get let's encrypt certificates")
	debugMode    = flag.Bool("debug", false, "Debug mode, disable let's encrypt, enable CORS and more logging")
)

func main() {

	// Parse the flags
	flag.Parse()

	// Initialize logger
	if *logFile != "" {
		log.SetFile(*logFile)
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
	log.Logger.Println("Main hostname is ", "https://"+os.Getenv("HOSTNAME")+":"+strconv.Itoa(*httpsPort))

	// Initializations
	jwt.Init(*debugMode)

	// Create the server
	rootMux := rootmux.CreateRootMux(*httpsPort, *appsFile, "web")

	// Serve locally with https on debug mode or with let's encrypt on production mode
	if *debugMode {
		// Init the hostname
		mocks.Init(*httpsPort)
		// Start a mock oauth2 server if debug mode is on
		mockOAuth2Port := ":8090"
		go http.ListenAndServe(mockOAuth2Port, mocks.CreateMockOAuth2())
		fmt.Println("Mock OAuth2 server Listening on: http://localhost" + mockOAuth2Port)
		// Start a mock API server if debug mode is on
		mockAPIPort := ":8091"
		go http.ListenAndServe(mockAPIPort, mocks.CreateMockAPI())
		fmt.Println("Mock API server Listening on: http://localhost" + mockAPIPort)
		log.Logger.Fatal(http.ListenAndServeTLS(":"+strconv.Itoa(*httpsPort), "./dev_certificates/localhost.crt", "./dev_certificates/localhost.key", log.Middleware(rootMux.Mux)))

	} else {
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache(*letsCacheDir),
			HostPolicy: rootMux.Policy,
		}

		server := &http.Server{
			Addr:    ":" + strconv.Itoa(*httpsPort),
			Handler: rootMux.Mux,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
			ReadTimeout:  30 * time.Minute, // in case of upload
			WriteTimeout: 5 * time.Hour,    // in case of download
			IdleTimeout:  120 * time.Second,
		}

		go http.ListenAndServe(":"+strconv.Itoa(*httpPort), certManager.HTTPHandler(nil))
		server.ListenAndServeTLS("", "")
	}
}
