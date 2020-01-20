package onlyoffice

import (
	"encoding/json"
	"net/http"
	"os"
	"text/template"
	"time"
)

// HandleOpen open the main onlyoffice  window
func HandleOpen(fullHostname string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		t, err := template.ParseFiles("web/onlyoffice/index.tmpl")
		if err != nil {
			http.Error(w, "could not open onlyoffice template: "+err.Error(), 500)
			return
		}
		title := os.Getenv("ONLYOFFICE_TITLE")
		if title == "" {
			title = "OnlyOffice"
		}
		p := struct {
			Title            string
			OnlyOfficeServer string
			Hostname         string
		}{title, os.Getenv("ONLYOFFICE_SERVER"), fullHostname}
		t.Execute(w, p)
	}
}

// HandleSaveCallback is the callback function wanted by onlyoffice to allow saving a document
// the body provides information on where to get the altered document, and the query provides information on where to put it
func HandleSaveCallback(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "the request method must be POST", 405)
		return
	}
	if req.Body == nil {
		http.Error(w, "the request must contain a body", 400)
		return
	}
	var bdy struct {
		Key        string `json:"key"`
		Status     int    `json:"status"`
		URL        string `json:"url"`
		Changesurl string `json:"changesurl"`
		History    struct {
			ServerVersion string `json:"serverVersion"`
			Changes       []struct {
				Created string `json:"created"`
				User    struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"user"`
			} `json:"changes"`
		} `json:"history"`
		Users   []string `json:"users"`
		Actions []struct {
			Type   int    `json:"type"`
			Userid string `json:"userid"`
		} `json:"actions"`
		Lastsave    time.Time `json:"lastsave"`
		Notmodified bool      `json:"notmodified"`
	}
	jsonErr := json.NewDecoder(req.Body).Decode(&bdy)
	if jsonErr != nil {
		http.Error(w, jsonErr.Error(), 400)
		return
	}
	// Case of document closed after editing
	if bdy.Status == 2 {
		// Get the binary content from url
		resp, err := http.Get(bdy.URL)
		if err != nil {
			http.Error(w, "could not get connect to onlyoffice document server", 400)
			return
		}
		defer resp.Body.Close()
		// PUT the content on the ressource gotten from the query
		ressource := req.URL.Query().Get("file") + "?token=" + req.URL.Query().Get("token")
		req, err := http.NewRequest("PUT", ressource, resp.Body)
		client := &http.Client{}
		_, err = client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"error\":0}"))
}
