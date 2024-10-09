package broker

import (
	"net/http"

	"github.com/eliona-smart-building-assistant/go-utils/log"
)

type webhookServer struct {
	secret string
	mux    *http.ServeMux
}

func newWebhookServer(secret string) *webhookServer {
	return &webhookServer{
		secret: secret,
		mux:    http.NewServeMux(),
	}
}

func (s *webhookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug("webhook", "Received request for URL: %s, Method: %s", r.URL.Path, r.Method)

	// TODO: Certificate check

	s.mux.ServeHTTP(w, r)
}

func (s *webhookServer) handleOntologyVersion(w http.ResponseWriter, r *http.Request) {
	//log.Debug("webhook", "Handled ontology version event: %+v", event)
}

func StartWebhookListener(secret string) {
	server := newWebhookServer(secret)

	server.mux.HandleFunc("/ontology-version", server.handleOntologyVersion)

	http.Handle("/", server)
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal("webhook", "Error starting server on port 8081: %v\n", err)
	}
}
