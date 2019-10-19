package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/asaskevich/govalidator"
	"github.com/mralanlee/url-shortener-api/db"
)

// Payload struct to place URL - initially thought to use for multi purposes
type Payload struct {
	Url string `json:"url"`
}

type InsertResponse struct {
	ShortenUrl string `json:"shortened_url"`
	Slug       string `json:"slug"`
}

type StatsResponse struct {
	Slug   string `json:"slug"`
	Source string `json:"source"`
	Vists  struct {
		LastTwentyFour int `json:"last_twenty_four"`
		Lifetime       int `json:"lifetime"`
	} `json:"visits"`
}

var hostname = os.Getenv("HOSTNAME")
var scheme = os.Getenv("SCHEME")

// Utility function for cehck if a URL is valid or not
func (p *Payload) IsValid() bool {
	if valid := govalidator.IsRequestURL(p.Url); !valid || p.Url == "" {
		return false
	}

	return true
}

func Shorten(client *db.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Request method not supported", http.StatusMethodNotAllowed)
			return
		}

		decoder := json.NewDecoder(r.Body)

		var payload Payload

		// Check if error parsing body or if URL is present
		if err := decoder.Decode(&payload); err != nil {
			http.Error(w, "Error reading request", http.StatusInternalServerError)
			return
		}

		// if URL key present, check if it's valid
		if !payload.IsValid() {
			http.Error(w, "Request contains invalid url", http.StatusBadRequest)
			return
		}

		slug := client.Insert(payload.Url)
		shortenedURL := fmt.Sprintf("%s://%s/%s", scheme, hostname, slug)
		insert := InsertResponse{
			ShortenUrl: shortenedURL,
			Slug:       slug,
		}

		response, jsonErr := json.Marshal(insert)

		if jsonErr != nil {
			http.Error(w, "Error shortening URL", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	})
}

func Stats(client *db.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check method
		if r.Method != http.MethodGet {
			http.Error(w, "Request method not supported", http.StatusMethodNotAllowed)
			return
		}

		req := r.URL.Query().Get("id")

		if req == "" {
			http.Error(w, "Missing url query parameter", http.StatusBadRequest)
			return
		}
		// need to return if query in db is invalid
		stats := client.ReadStats(req)

		// TODO: need to check if stats is empty
		result, err := json.Marshal(stats)

		if err != nil {
			http.Error(w, "Error fetching stats", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(result)
	})
}

func Redirect(client *db.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Request method not supported", http.StatusMethodNotAllowed)
			return
		}
		slug := r.URL.Path[len("/"):]

		// need to handle null
		source := client.FindSlug(slug)

		if source == "" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		http.Redirect(w, r, source, 301)

		// need to increment
		client.Increment(slug)
	})
}
