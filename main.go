// Please have a look at our cloud service immudb Vault and build a simple application (Backend + Frontend) around it with the following requirements:
//
//	Application is storing accounting information within immudb Vault with the following structure:
//	    account number (unique), account name, iban, address, amount, type (sending, receiving)
//	Application has an API to add and retrieve accounting information
//	Application has a frontend that displays accounting information and allows to create new records.
//
// The solution should:
//
//	Have a readme
//	Have a documented API
//	Have docker-compose so it is easy to run.
//
// Resources:
// immudb Vault documentation: https://vault.immudb.io/docs/
// API reference: https://vault.immudb.io/docs/api/v1
package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

var apiKey, apiROKey, apiURL, addr string

type Account struct {
	Number   string `json:"number"` // unique
	Name     string `json:"name"`
	IBAN     string `json:"iban"`
	Address  string `json:"address"`
	Amount   float64 `json:"amount"`
	Type     string `json:"type"`
}

const (
	typeSending   = "sending"
	typeReceiving = "receiving"
)

type AccountManager struct {
    mux *http.ServeMux
	addr string
	server *http.Server
	timeout time.Duration
}

func NewAccountManager(addr string, timeout time.Duration) *AccountManager {
	am := &AccountManager{
		addr: addr,
		timeout: timeout,
	}

	am.mux = http.NewServeMux()
	am.mux.HandleFunc("/", am.handlerFrontend)
	am.mux.HandleFunc("/api/add", am.handlerAdd)
	am.mux.HandleFunc("/api/get", am.handlerGet)

	am.server = &http.Server{
		Addr: addr,
		Handler: am.mux,
		ReadTimeout: 10 * time.Second,
		MaxHeaderBytes: 4096,
	}

	ok, err := am.collectionExists()
	if err != nil {
		if err == ErrCollectionNotFound {
			if err := am.collectionCreate(); err != nil {
				log.Printf("Failed to create collection: %v", err)
				return nil
			}
			log.Println("Collection created")

		} else {
			log.Printf("Failed to check collection existence: %v", err)
			return nil
		}

	} else if !ok {
		log.Println("Collection does not exist")
		return nil

	} else {
		log.Println("Collection exists")
	}
	return am
}

func (am *AccountManager) handlerFrontend(w http.ResponseWriter, r *http.Request) {
	// tmpl, err := template.ParseFiles("index.html")
	tmpl, err := template.New("index").Parse(indexHTML)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (am *AccountManager) handlerAdd(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to add account")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		log.Printf(`Method "%v" not allowed`, r.Method)
		return
	}

	// ------------------------------------------------------------------------- parse account
	var account Account
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&account); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		log.Printf("Invalid request: %v\n", err)
		return
	}

	// ------------------------------------------------------------------------- validate
	if !validateIBAN(account.IBAN) ||
	   !validateNumber(account.Number) ||
	   !validateName(account.Name) ||
	   !validateAddress(account.Address) ||
	   !validateAmount(account.Amount) ||
	   !validateAccountType(account.Type) {

		http.Error(w, "invalid account data", http.StatusBadRequest)
		log.Println("Invalid account data")
		return
	}

	// ------------------------------------------------------------------------- check exists
	exists, err := am.accountExists(account.Number)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("accountExists: %v", err)
		return
	}

	if exists {
		http.Error(w, "account already exists", http.StatusConflict)
		log.Printf("Account already exists")
		return
	}
	// ------------------------------------------------------------------------- add account to Vault
	if err := am.accountAdd(&account); err != nil {
		http.Error(w, "failed to add account", http.StatusInternalServerError)
		log.Printf("Failed to add account: %v\n", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	log.Println("Account added successfully")
}

func (am *AccountManager) handlerGet(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		log.Printf(`Method "%v" not allowed`, r.Method)
		return
	}

	var query apiQuery

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&query); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		log.Printf("Invalid request: %v\n", err)
		return
	}

	var accounts []*Account
	var err error

	if query.Number != "" {
		accounts, err = am.accountGet(query.Number)

	} else if query.Type != "" {
		accounts, err = am.accountGet(query.Type)

	} else {
		http.Error(w, "invalid query", http.StatusBadRequest)
		log.Println("Invalid query")
		return
	}

	if err != nil {
		http.Error(w, "failed to retrieve accounts", http.StatusInternalServerError)
		log.Printf("Failed to retrieve accounts: %v\n", err)
		return
	}

	if len(accounts) == 0 {
		http.Error(w, "account not found", http.StatusNotFound)
		log.Println("Account not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(accounts); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		log.Printf("Failed to encode response: %v\n", err)
		return
	}
	log.Println("Accounts retrieved successfully")
}

func (am *AccountManager) Start() (err error) {
	log.Printf("Server starting on port %s\n", am.addr)
	return am.server.ListenAndServe()
}

func getEnvironment() bool {
	addr = os.Getenv("PORT")
	if addr == "" {
		log.Printf("-ERR: PORT is not set")
		return false
	}
	addr = ":" + addr

	apiKey = os.Getenv("API_KEY")
	if apiKey == "" {
		log.Printf("-ERR: API_KEY is not set")
		return false
	}

	apiROKey = os.Getenv("API_RO_KEY")
	if apiROKey == "" {
		log.Printf("-ERR: API_RO_KEY is not set")
		return false
	}

	apiURL = os.Getenv("API_URL")
	if apiURL == "" {
		log.Printf("-ERR: API_URL is not set")
		return false
	}
	return true
}

func main() {
	if !getEnvironment() {
		log.Fatal("Failed to get environment")
	}
	log.Printf("Environment read successfully")

	am := NewAccountManager(addr, 5 * time.Second)
	if am == nil {
		log.Fatalf("Failed to create AccountManager")
	}

	if err := am.Start(); err != nil {
		log.Fatalf("Failed to start AccountManager: %v", err)
	}
}
