package broker

import (
	"io"
	"net/http"

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

	// TODO: Certificate check

	s.mux.ServeHTTP(w, r)
}

func (s *webhookServer) handleOntologyVersion(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	log.Debug("webhook", "Received ontology request headers: %+v", r.Header)
	log.Debug("webhook", "Received ontology request body: %s", body)

	// TODO: Actually implement version parsing once we know the fromat of the data.
}

func StartWebhookListener() {
	server := newWebhookServer()

	server.mux.HandleFunc("/ontology-version", server.handleOntologyVersion)

	http.Handle("/", server)
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal("webhook", "Error starting server on port 8081: %v\n", err)
	}
}
