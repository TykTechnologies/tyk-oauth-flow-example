package main

import (
	log "github.com/Sirupsen/logrus"
	"html/template"
	"net/http"
)

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "tmpl/index.html")
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmplVal := make(map[string]string)
	tmplVal["ClientID"] = r.FormValue("client_id")
	tmplVal["ResponseType"] = r.FormValue("response_type")
	tmplVal["RedirectURI"] = r.FormValue("redirect_uri")
	tmplVal["State"] = r.FormValue("state")

	t, _ := template.ParseFiles("tmpl/login.html")
	t.Execute(w, tmplVal)
}

func approvedHandler(w http.ResponseWriter, r *http.Request) {
	var redirect_uri = r.FormValue("redirect_uri")
	var responseType = r.FormValue("response_type")
	var clientId = r.FormValue("client_id")
	var State = r.FormValue("state")

	thisResponse, rErr := RequestOAuthToken(APIlistenPath,
		redirect_uri, responseType, clientId, "", orgID, policyID, BaseAPIID, State)

	if rErr != nil {
		log.Error(rErr)
		http.Error(w, "Error!", 500)
	}
	http.Redirect(w, r, thisResponse.RedirectTo, 301)
}

func finalHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "tmpl/final.html")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/approved", approvedHandler)
	mux.HandleFunc("/final", finalHandler)
	mux.HandleFunc("/", index)
	log.Info("Listening on :8000")
	http.ListenAndServe(":8000", mux)
}
