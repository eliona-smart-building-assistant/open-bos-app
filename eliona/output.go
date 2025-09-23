package eliona

import (
	"time"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v3"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/gorilla/websocket"
)

// ListenForOutputChanges on assets (only output attributes). Returns a channel with all changes.
func ListenForOutputChanges(apiKey string) (chan api.Data, error) {
	outputs := make(chan api.Data)
	go http.ListenWebSocketWithReconnectAlways(func() (*websocket.Conn, error) {
		return newDataWebsocket(apiKey)
	}, time.Duration(0), outputs)
	return outputs, nil
}

func newDataWebsocket(apiKey string) (*websocket.Conn, error) {
	return http.NewWebSocketConnectionWithApiKey(common.Getenv("API_ENDPOINT", "")+"/data-listener?dataSubtype=output", "X-API-Key", apiKey)
}

// ListenForAlarmChanges on assets (only output attributes). Returns a channel with all changes.
func ListenForAlarmChanges(apiKey string) (chan api.Alarm, error) {
	outputs := make(chan api.Alarm)
	go http.ListenWebSocketWithReconnectAlways(func() (*websocket.Conn, error) {
		return newAlarmWebsocket(apiKey)
	}, time.Duration(0), outputs)
	return outputs, nil
}

func newAlarmWebsocket(apiKey string) (*websocket.Conn, error) {
	return http.NewWebSocketConnectionWithApiKey(common.Getenv("API_ENDPOINT", "")+"/alarm-listener", "X-API-Key", apiKey)
}
