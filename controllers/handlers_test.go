package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/mralanlee/url-shortener-api/db"
)

type TestPayload struct {
	Url    string `json:"url"`
	Random string `json:"random"`
}

var shortenErrorStatusTests = []struct {
	Method   string
	Payload  TestPayload
	Expected int
}{
	{"GET", TestPayload{}, http.StatusMethodNotAllowed},
	{"POST", TestPayload{Url: "asdf"}, http.StatusBadRequest},
	{"POST", TestPayload{Random: "random"}, http.StatusBadRequest},
	{"POST", TestPayload{Url: "https://www.google.com"}, http.StatusOK},
}

var statsErrorStatusTests = []struct {
	Method   string
	Query    string
	Expected int
}{
	{"GET", "key?=asdf", http.StatusBadRequest},
	{"POST", "?id=hello", http.StatusMethodNotAllowed},
	{"GET", "?id=", http.StatusOK},
}

var client = &db.Client{
	Sql: db.Init(),
}

func shortener(method string, sample TestPayload) *httptest.ResponseRecorder {
	handler := Shorten(client)
	rr := httptest.NewRecorder()

	payload, _ := json.Marshal(sample)
	req, err := http.NewRequest(method, "/", bytes.NewBuffer(payload))

	if err != nil {
		log.Fatal(err)
	}

	handler.ServeHTTP(rr, req)

	return rr
}

func TestShortenEndpoint(t *testing.T) {
	for _, tt := range shortenErrorStatusTests {
		testcase := shortener(tt.Method, tt.Payload)
		if status := testcase.Code; status != tt.Expected {
			t.Errorf("handler returned wrong status code with message: %s, got %v want %v", testcase.Body.String(), status, tt.Expected)
		}
	}
}

func TestStatsEndpoint(t *testing.T) {
	for _, tt := range statsErrorStatusTests {
		handler := Stats(client)
		rr := httptest.NewRecorder()

		var requestEndpoint string

		if tt.Expected != http.StatusOK {
			requestEndpoint = fmt.Sprintf("/%s", tt.Query)
		} else {
			success := shortenErrorStatusTests[len(shortenErrorStatusTests)-1]
			resp := shortener(success.Method, success.Payload)
			result := new(InsertResponse)
			jsonErr := json.Unmarshal(resp.Body.Bytes(), &result)

			if jsonErr != nil {
				t.Fatal(jsonErr)
			}

			requestEndpoint = fmt.Sprintf("/%s%s", tt.Query, result.Slug)
		}
		req, err := http.NewRequest(tt.Method, requestEndpoint, nil)

		if err != nil {
			t.Fatal(err)
		}

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != tt.Expected {
			t.Errorf("handler returned wrong status code with message: %s, got %v want %v", rr.Body.String(), status, tt.Expected)
		}
	}
}
