package main

import (
	"fmt"
	"net/http"
	"os"
	"encoding/json"
	"strings"
	"time"
	"github.com/thedevsaddam/renderer"
)

const clientID = "1729e8ebab6944327671308fe14e518deb9a1bd9186eb4268a571efe64b03f8c"
const clientSecret = "ca65630a8809f765712774bf86ebb530389e05c09f0c9f2547a501011caaed40"

type SaleList struct {
	_attributes struct {
		Count  string `json:"count"`
		Limit  string `json:"limit"`
		Offset string `json:"offset"`
	} `json:"@attributes"`
	Sale []struct {
		Archived              string `json:"archived"`
		Balance               string `json:"balance"`
		CalcAvgCost           string `json:"calcAvgCost"`
		CalcDiscount          string `json:"calcDiscount"`
		CalcFIFOCost          string `json:"calcFIFOCost"`
		CalcNonTaxable        string `json:"calcNonTaxable"`
		CalcPayments          string `json:"calcPayments"`
		CalcSubtotal          string `json:"calcSubtotal"`
		CalcTax1              string `json:"calcTax1"`
		CalcTax2              string `json:"calcTax2"`
		CalcTaxable           string `json:"calcTaxable"`
		CalcTotal             string `json:"calcTotal"`
		Change                string `json:"change"`
		CompleteTime          string `json:"completeTime"`
		Completed             string `json:"completed"`
		CreateTime            string `json:"createTime"`
		CustomerID            string `json:"customerID"`
		DiscountID            string `json:"discountID"`
		DiscountPercent       string `json:"discountPercent"`
		DisplayableSubtotal   string `json:"displayableSubtotal"`
		DisplayableTotal      string `json:"displayableTotal"`
		EmployeeID            string `json:"employeeID"`
		EnablePromotions      string `json:"enablePromotions"`
		IsTaxInclusive        string `json:"isTaxInclusive"`
		QuoteID               string `json:"quoteID"`
		ReceiptPreference     string `json:"receiptPreference"`
		ReferenceNumber       string `json:"referenceNumber"`
		ReferenceNumberSource string `json:"referenceNumberSource"`
		RegisterID            string `json:"registerID"`
		SaleID                string `json:"saleID"`
		ShipToID              string `json:"shipToID"`
		ShopID                string `json:"shopID"`
		Tax1Rate              string `json:"tax1Rate"`
		Tax2Rate              string `json:"tax2Rate"`
		TaxCategoryID         string `json:"taxCategoryID"`
		TaxTotal              string `json:"taxTotal"`
		TicketNumber          string `json:"ticketNumber"`
		TimeStamp             string `json:"timeStamp"`
		Total                 string `json:"total"`
		TotalDue              string `json:"totalDue"`
		UpdateTime            string `json:"updateTime"`
		Voided                string `json:"voided"`
	} `json:"Sale"`
}

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int32 `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType string `json:"token_type"`
	Scope string `json:"scope"`
}

type AccountParams struct {
	Account struct{
		AccountID string `json:"accountID"`
		AccountName string `json:"name"`
	}
}

var t OAuthAccessResponse
var acc AccountParams
var expiredTime int32
var rnd *renderer.Render

func getToken(w http.ResponseWriter, r *http.Request) {

		getCodeBody := fmt.Sprintf("https://cloud.lightspeedapp.com/oauth/authorize.php?response_type=code&client_id=%s&scope=%s", clientID, "employee:admin_void_sale")
		w.Header().Set("Location", getCodeBody)
		w.WriteHeader(http.StatusFound)
}

func getOauth(w http.ResponseWriter, r *http.Request) {
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

	/*------------------------------------------------GET AccessID------------------------------------------------*/

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

	if acc.Account.AccountID != "" {
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusFound)
	}
}

func refreshToken() {
	body := strings.NewReader(fmt.Sprintf(`refresh_token=%s&client_id=%s&client_secret=%s&grant_type=refresh_token`, t.RefreshToken, clientID, clientSecret))

	req, err := http.NewRequest("POST", "https://cloud.lightspeedapp.com/oauth/access_token.php", body)

	if err != nil {
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
		//w.WriteHeader(http.StatusBadRequest)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send out the HTTP request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
		//w.WriteHeader(http.StatusInternalServerError)
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		//w.WriteHeader(http.StatusBadRequest)
	}
	return
}

func getSales(w http.ResponseWriter, r *http.Request)  {
	if t.AccessToken == "" {
		w.Header().Set("Location", "/auth")
		w.WriteHeader(http.StatusFound)
		return
	} else if (expiredTime < int32(time.Now().Unix())) {
		refreshToken()
	}

	req, err := http.NewRequest("POST", "https://api.lightspeedapp.com/API/Account/" + acc.Account.AccountID + "/Sale.json", nil)

	if err != nil {
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}
	bearer := fmt.Sprintf("Bearer %s", t.AccessToken)
	req.Header.Set("Authorization", bearer)

	// Send out the HTTP request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer res.Body.Close()

	var sale SaleList

	if err := json.NewDecoder(res.Body).Decode(&sale); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	fmt.Println("Sale ", len(sale.Sale))

	rnd.HTML(w, http.StatusOK, "index", sale.Sale)
}

func main() {
	http.HandleFunc("/", getSales)
	http.HandleFunc("/auth", getToken)
	http.HandleFunc("/oauth/redirect", getOauth)

	http.ListenAndServe("https//:8888", nil)
}

func init() {
	opts := renderer.Options{
		ParseGlobPattern: "./public/*.html",
	}
	rnd = renderer.New(opts)
}
