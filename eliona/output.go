package eliona

import (
	"time"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/gorilla/websocket"
)

// ListenForOutputChanges on assets (only output attributes). Returns a channel with all changes.
func ListenForOutputChanges() (chan api.Data, error) {
	outputs := make(chan api.Data)
	go http.ListenWebSocketWithReconnectAlways(newDataWebsocket, time.Duration(0), outputs)
	return outputs, nil
}

func newDataWebsocket() (*websocket.Conn, error) {
	return http.NewWebSocketConnectionWithApiKey(common.Getenv("API_ENDPOINT", "")+"/data-listener?dataSubtype=output", "X-API-Key", common.Getenv("API_TOKEN", ""))
}

// ListenForAlarmChanges on assets (only output attributes). Returns a channel with all changes.
func ListenForAlarmChanges() (chan api.Alarm, error) {
	outputs := make(chan api.Alarm)
	go http.ListenWebSocketWithReconnectAlways(newAlarmWebsocket, time.Duration(0), outputs)
	return outputs, nil
}

func newAlarmWebsocket() (*websocket.Conn, error) {
	return http.NewWebSocketConnectionWithApiKey(common.Getenv("API_ENDPOINT", "")+"/alarm-listener", "X-API-Key", common.Getenv("API_TOKEN", ""))
}
