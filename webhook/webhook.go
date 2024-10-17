package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"open-bos/app"
	"regexp"
	"strconv"
	"time"

	"github.com/eliona-smart-building-assistant/go-utils/log"
)

type webhookServer struct {
	mux *http.ServeMux
}

func newWebhookServer() *webhookServer {
	return &webhookServer{
		mux: http.NewServeMux(),
	}
}

func (s *webhookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug("webhook", "Received request for URL: %s, Method: %s", r.URL.Path, r.Method)

	configID, err := parseConfigIDFromPath(r.URL.Path)
	if err != nil {
		log.Warn("webhook", "Invalid URL path, missing or invalid config ID: %s", r.URL.Path)
		http.Error(w, "Invalid URL path, missing or invalid config ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, "configID", configID)
	r = r.WithContext(ctx)

	r.URL.Path = removeConfigIDFromPath(r.URL.Path)

	s.mux.ServeHTTP(w, r)
}

func (s *webhookServer) handleOntologyVersion(w http.ResponseWriter, r *http.Request) {
	configID := r.Context().Value("configID").(int64)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	log.Debug("webhook", "Received ontology request headers: %+v", r.Header)
	log.Debug("webhook", "Request body: %s", body)
	log.Debug("webhook", "Method: %s", r.Method)
	log.Debug("webhook", "Config ID: %d", configID)

	// TODO: Implement version parsing once we know the format of the data.

	app.CollectConfigData(configID)
}

func (s *webhookServer) handleLivedataUpdate(w http.ResponseWriter, r *http.Request) {
	configID := r.Context().Value("configID").(int64)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	log.Debug("webhook", "Received data request headers: %+v", r.Header)
	log.Debug("webhook", "Request body: %s", body)
	log.Debug("webhook", "Method: %s", r.Method)
	log.Debug("webhook", "Config ID: %d", configID)

	type LiveData struct {
		ID         string      `json:"Id"`
		IsProperty bool        `json:"IsProperty"`
		TimeStamp  string      `json:"TimeStamp"`
		Quality    string      `json:"Quality"`
		Value      interface{} `json:"Value"`
	}

	var liveData []LiveData
	if err := json.Unmarshal(body, &liveData); err != nil {
		log.Error("webhook", "Failed to parse request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for _, data := range liveData {
		// Timestamps are always in UTC - see docs
		layout := "02/01/2006 15:04:05"
		timestamp, err := time.ParseInLocation(layout, data.TimeStamp, time.UTC)
		if err != nil {
			log.Warn("webhook", "Invalid timestamp format for ID %s: %v", data.ID, err)
			continue
		}

		log.Debug("webhook", "Received data for ID %s: IsProperty=%v, TimeStamp=%v, Quality=%s, Value=%v",
			data.ID, data.IsProperty, timestamp, data.Quality, data.Value)

		// TODO: Use the update
	}

	w.WriteHeader(http.StatusOK)
}

func parseConfigIDFromPath(path string) (int64, error) {
	// Matches "/{configID}/rest-of-path"
	re := regexp.MustCompile(`^/(\d+)/`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 2 {
		return 0, fmt.Errorf("config ID not found in path")
	}
	return strconv.ParseInt(matches[1], 10, 64)
}

func removeConfigIDFromPath(path string) string {
	re := regexp.MustCompile(`^/\d+/`)
	return re.ReplaceAllString(path, "/")
}

func StartWebhookListener() {
	server := newWebhookServer()

	server.mux.HandleFunc("/ontology-version", server.handleOntologyVersion)
	server.mux.HandleFunc("/ontology-livedata", server.handleLivedataUpdate)

	http.Handle("/", server)
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal("webhook", "Error starting server on port 8081: %v\n", err)
	}
}
