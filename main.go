package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const clientID = "1729e8ebab6944327671308fe14e518deb9a1bd9186eb4268a571efe64b03f8c"
const clientSecret = "ca65630a8809f765712774bf86ebb530389e05c09f0c9f2547a501011caaed40"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		getCodeBody := fmt.Sprintf("https://cloud.lightspeedapp.com/oauth/authorize.php?response_type=code&client_id=%s&scope=%s", clientID, "employee:inventory+employee:reports")
		w.Header().Set("Location", getCodeBody)
		w.WriteHeader(http.StatusFound)
	})

	http.HandleFunc("/oauth/redirect", func(w http.ResponseWriter, r *http.Request) {
		// First, we need to get the value of the `code` query param
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not parse query: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		code := r.FormValue("code")

		fmt.Println(code)

		body := strings.NewReader(fmt.Sprintf(`client_id=%s&client_secret=%s&code=%s&grant_type=authorization_code`, clientID, clientSecret, code))

		req, err := http.NewRequest("POST", "https://cloud.lightspeedapp.com/oauth/access_token.php", body)

		if err != nil {
			fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Send out the HTTP request
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		defer res.Body.Close()

		// Parse the request body into the `OAuthAccessResponse` struct
		var t OAuthAccessResponse
		if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
			fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}

		// Finally, send a response to redirect the user to the "welcome" page
		// with the access token
		//w.Header().Set("Location", "/welcome.html?access_token="+t.AccessToken)
		//w.WriteHeader(http.StatusFound)
	})

	http.ListenAndServe(":8888", nil)
}

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType string `json:"token_type"`
	Scope string `json:"scope"`
}
