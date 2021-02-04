package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

const clientID = "df0099cc8df7d929d0e2"

const clientSecret = "d422926df16d50d62c662cdaa7cadbf962cff402"

var redirect_url = "https://github.com/login/oauth/authorize?client_id=" + clientID + "&redirect_uri=http://localhost:8080/callback"

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}

func loginWithGithub(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, redirect_url, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	httpClient := http.Client{}
	log.Println("Inside login method")
	code := r.FormValue("code")
	reqURL := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", clientID, clientSecret, code)
	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	req.Header.Set("accept", "application/json")

	res, err := httpClient.Do(req)

	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	var token OAuthAccessResponse
	if err := json.NewDecoder(res.Body).Decode(&token); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	cookie := http.Cookie{Name: "token", Value: token.AccessToken, Expires: time.Now().Add(60 * time.Second), HttpOnly: true}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, "/", 301)

}

func serveHome(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("token")
	if err != nil {
		log.Println("no token found")
		http.Redirect(w, r, "/login/github", http.StatusTemporaryRedirect)
	} else {
		log.Println("token found")
		fmt.Fprint(w, token.Value)
	}

}
func main() {
	router := mux.NewRouter()
	router.HandleFunc("/login/github", loginWithGithub)
	router.HandleFunc("/callback", handleCallback)
	router.HandleFunc("/", serveHome)
	http.ListenAndServe(":8080", router)
}
