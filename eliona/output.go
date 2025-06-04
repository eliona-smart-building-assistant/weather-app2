package eliona

import (
	"time"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/gorilla/websocket"
)

// ListenForPropertyChanges on assets (only output attributes). Returns a channel with all changes.
func ListenForPropertyChanges() (chan api.Data, error) {
	outputs := make(chan api.Data)
	go http.ListenWebSocketWithReconnectAlways(newWebsocket, time.Duration(0), outputs)
	return outputs, nil
}

func newWebsocket() (*websocket.Conn, error) {
	return http.NewWebSocketConnectionWithApiKey(common.Getenv("API_ENDPOINT", "")+"/data-listener?dataSubtype=property", "X-API-Key", common.Getenv("API_TOKEN", ""))
}
