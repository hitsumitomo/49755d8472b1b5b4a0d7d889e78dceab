package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	contentType = "application/json"
	countRequest = `{
  "query": {
    "expressions": [
      {
        "fieldComparisons": [
          {
            "field": "number",
            "operator": "EQ",
            "value": "%s"
          }
        ]
      }
    ]
  }
}`
	accountNumberRequest = `{
  "page": 1,
  "perPage": 1,
  "query": {
    "expressions": [
      {
        "fieldComparisons": [
          {
            "field": "number",
            "operator": "EQ",
            "value": "%s"
          }
        ]
      }
    ]
  }
}`
	accountTypeRequest = `{
  "page": 1,
  "perPage": 100,
  "query": {
    "expressions": [
      {
        "fieldComparisons": [
          {
            "field": "type",
            "operator": "EQ",
            "value": "%s"
          }
        ]
      }
    ]
  }
}`
	createCollection = `{
  "idFieldName": "_id",
  "fields": [
    {
      "name": "number",
      "type": "STRING"
    },
    {
      "name": "type",
      "type": "STRING"
    }
  ],
  "indexes": [
    {
      "fields": [
        "number"
      ],
      "isUnique": true
    },
    {
      "fields": [
        "type"
      ],
      "isUnique": false
    }
  ]
}`
)

var (
	ErrCollectionNotFound = errors.New("collection not found")
)

type VaultResponse struct {
	Page      int `json:"page"`
	PerPage   int `json:"perPage"`
	Revisions []struct {
		Document struct {
			Account
		} `json:"document"`
		Revision      string `json:"revision"`
		TransactionID string `json:"transactionId"`
	} `json:"revisions"`
	SearchID string `json:"searchId"`
}

type VaultCountResponse struct {
	Collection string `json:"collection"`
	Count      int    `json:"count"`
}

type apiQuery struct {
	Number string `json:"number"`
	Type   string `json:"type"`
}

func (am *AccountManager) vaultRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, apiURL + endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", contentType)
	req.Header.Set("Content-Type", contentType)
	if method == http.MethodPut {
		req.Header.Set("X-API-Key", apiKey)
	} else {
		req.Header.Set("X-API-Key", apiROKey)
	}

	client := &http.Client{
		Timeout: am.timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (am *AccountManager) accountAdd(account *Account) (err error) {
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(account)
	if err != nil {
		return err
	}

	resp, err := am.vaultRequest(http.MethodPut, "/document", b)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add account: status: %s", resp.Status)
	}
	return nil
}

func (am *AccountManager) accountGet(request string) (accounts []*Account, err error) {
	b := &bytes.Buffer{}

	if request == typeSending || request == typeReceiving {
		fmt.Fprintf(b, accountTypeRequest, request)
	} else {
		fmt.Fprintf(b, accountNumberRequest, request)
	}

	resp, err := am.vaultRequest(http.MethodPost, "/documents/search", b)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve account: %s", resp.Status)
	}

	var result VaultResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Revisions) == 0 {
		return
	}

	for _, revision := range result.Revisions {
		account := &revision.Document.Account
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (am *AccountManager) collectionExists() (bool, error) {
	resp, err := am.vaultRequest(http.MethodGet, "", nil)
	if err != nil {
		return false, err
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil

	} else if resp.StatusCode == http.StatusNotFound {
		return false, ErrCollectionNotFound
	}
	return false, nil
}

func (am *AccountManager) collectionCreate() error {
	resp, err := am.vaultRequest(http.MethodPut, "", bytes.NewBufferString(createCollection))
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create collection: %s", resp.Status)
	}
	return nil
}

func (am *AccountManager) accountExists(number string) (bool, error) {
	b := &bytes.Buffer{}
	fmt.Fprintf(b, countRequest, number)
	resp, err := am.vaultRequest(http.MethodPost, "/documents/count", b)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to retrieve accounts: %s", resp.Status)
	}

	var result VaultCountResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}
	return result.Count > 0, nil
}
