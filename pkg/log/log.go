package log

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	maxminddb "github.com/oschwald/maxminddb-golang"
)

type cache struct {
	mux     sync.Mutex
	last    time.Time
	content map[string]string
}

// Logger represents a standard logger sets up for this application usage
var (
	Logger  *log.Logger
	ipcache = cache{
		last:    time.Now(),
		content: make(map[string]string),
	}
	ipDbLocation = "./ipgeodatabase/GeoLite2-City.mmdb"
	f            *os.File
)

func init() {
	// Initialize logger
	Logger = log.New(os.Stdout, "", log.LstdFlags)
}

// SetFile set a file to log to instead of standard output
func SetFile(file string) {
	var err error
	f, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	wrt := io.MultiWriter(os.Stdout, f)
	Logger.SetOutput(wrt)
}

// CloseFile close the log file on exit
func CloseFile() {
	f.Close()
}

// GetCityAndCountryFromRequest returns a string containing the city and the contry where the request is from
// If the city is fetched from the built-in memory cache, it is NOT suffixed by a dot (.)
func GetCityAndCountryFromRequest(req *http.Request) string {
	// If the request remote adress is local return "localhost"
	if req.RemoteAddr == "" || strings.HasPrefix(req.RemoteAddr, "[::1]") || strings.HasPrefix(req.RemoteAddr, "127.0.0.1") {
		return "localhost"
	}

	// Lock the cache
	ipcache.mux.Lock()
	defer ipcache.mux.Unlock()
	// Check if the cache is to old or to big
	if time.Now().After(ipcache.last.Add(time.Hour*24)) || len(ipcache.content) > 1000 {
		// If so reset the cache
		ipcache.last = time.Now()
		ipcache.content = make(map[string]string)
	}

	// Get ip from remote address
	address := strings.Split(req.RemoteAddr, ":")[0]

	// First check if the ip is in memory cache
	locFromCache, ok := ipcache.content[address]
	if ok {
		return locFromCache
	}

	// If not open the maxmind database, search the ip and update the cache
	db, err := maxminddb.Open(ipDbLocation)
	if err != nil {
		Logger.Fatal(err)
	}
	defer db.Close()

	ip := net.ParseIP(address)

	if ip == nil {
		return "ip could not be parsed"
	}

	var record struct {
		City struct {
			Names map[string]string `maxminddb:"names"`
		} `maxminddb:"city"`
		Country struct {
			Names map[string]string `maxminddb:"names"`
		} `maxminddb:"country"`
	}

	err = db.Lookup(ip, &record)
	if err != nil {
		Logger.Fatal(err)
	}
	if record.Country.Names["fr"] == "" {
		return "ip not found"
	}
	ipFromDB := fmt.Sprintf("%v, %v", record.City.Names["fr"], record.Country.Names["fr"])
	ipcache.content[address] = ipFromDB
	return ipFromDB + "."
}
