package di

import (
	"darbelis.eu/wshub/internal/messages"
	"darbelis.eu/wshub/ws"
)

var wsHubInstance *ws.Hub = nil

func GetWsHub() *ws.Hub {
	if wsHubInstance == nil {
		wsHubInstance = ws.NewHub("common", messages.Encode, messages.Decode)
	}

	return wsHubInstance
}
