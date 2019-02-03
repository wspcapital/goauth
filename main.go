package main

import (
	"fmt"
	"net/http"
	"os"
	"encoding/json"
	"strings"
	"time"
)

const clientID = "1729e8ebab6944327671308fe14e518deb9a1bd9186eb4268a571efe64b03f8c"
const clientSecret = "ca65630a8809f765712774bf86ebb530389e05c09f0c9f2547a501011caaed40"

func getToken() {
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

		if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
			fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		fmt.Println("t.AccessToken ", t.AccessToken)

		expiredTime = int32(time.Now().Unix()) + t.ExpiresIn

		// Finally, send a response to redirect the user to the "welcome" page
		// with the access token
		//w.Header().Set("Location", "/welcome.html?access_token="+t.AccessToken)
		//w.WriteHeader(http.StatusFound)

		req_acc, err := http.NewRequest("GET", "https://api.lightspeedapp.com/API/Account.json", nil)
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		bearer := fmt.Sprintf("Bearer %s", t.AccessToken)
		req_acc.Header.Set("Authorization", bearer)

		resp, err := http.DefaultClient.Do(req_acc)
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&acc); err != nil {
			fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		fmt.Println("acc.AccountID ", acc.Account.AccountID)

		/*req_acc, err := http.NewRequest("GET", "https://api.lightspeedapp.com/API/Account.json", nil)
		if err != nil {
			// handle err
		}
		req_acc.Header.Set("Authorization", "Bearer 80c42e112f3021d5388005134b850002ae7004e6")

		resp, err := http.DefaultClient.Do(req_acc)
		if err != nil {
			// handle err
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&acc); err != nil {
			fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
			//w.WriteHeader(http.StatusBadRequest)
		}
		fmt.Println("acc.AccountID ", acc.Account.AccountID)*/
	})
}

func refreshToken() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// First, we need to get the value of the `code` query param
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not parse query: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}

		body := strings.NewReader(fmt.Sprintf(`refresh_token=%s&client_secret=%s&client_id=%s&grant_type=refresh_token`, t.RefreshToken, clientSecret, clientID))

		req, err := http.NewRequest("POST", "https://cloud.lightspeedapp.com/oauth/access_token.php", body)

		if err != nil {
			fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}

		// Send out the HTTP request
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		defer res.Body.Close()

		if err := json.NewDecoder(res.Body).Decode(&rt); err != nil {
			fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		fmt.Println("rt.AccessToken ", rt.AccessToken)
	})
}

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int32 `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType string `json:"token_type"`
	Scope string `json:"scope"`
}

type OAuthRefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int32 `json:"expires_in"`
	TokenType string `json:"token_type"`
	Scope string `json:"scope"`
}

type AccountParams struct {
	Account struct{
		AccountID string `json:"accountID"`
		AccountName string `json:"name"`
	}
 }

type SaleLine struct {
	itemID int
	unitQuantity int
}

type SaleStruct struct {
	employeeID int
	registerID int
	shopID int
	customerID int
	completed bool
	SaleLines struct {
		SaleLine [] SaleLine
	}
	SalePayments struct {
		SalePayment struct {
			amount float64
			paymentTypeID int
		}
	}
}

var t OAuthAccessResponse
var rt OAuthRefreshResponse
var acc AccountParams
var expiredTime int32
var s SaleStruct


func main() {
	if t.AccessToken == "" {
		getToken()
		fmt.Println("B-", int32(time.Now().Unix()))
	} else if (expiredTime < int32(time.Now().Unix())) {
		refreshToken()
	}

	http.HandleFunc("/sale", func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("/sale")
		fmt.Println("acc: ", acc.Account.AccountID)

		req, err := http.NewRequest("POST", "https://api.lightspeedapp.com/API/Account/" + acc.Account.AccountID + "/Sale.json", nil)

		if err != nil {
			fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}

		req.Header.Set("Authorization", "Bearer " + t.AccessToken)

		// Send out the HTTP request
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		defer res.Body.Close()

		if err := json.NewDecoder(res.Body).Decode(&s); err != nil {
			fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		fmt.Println("body: ", res.Body)
		fmt.Println("s: ", s)
	})


	// Generated by curl-to-Go: https://mholt.github.io/curl-to-go

	/*req, err := http.NewRequest("GET", "https://api.lightspeedapp.com/API/Account.json", nil)
	if err != nil {
		// handle err
	}
	req.Header.Set("Authorization", "Bearer 80c42e112f3021d5388005134b850002ae7004e6")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&acc); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		//w.WriteHeader(http.StatusBadRequest)
	}
	fmt.Println("acc.AccountID ", acc.Account.AccountID)*/

	http.ListenAndServe(":8888", nil)
}


