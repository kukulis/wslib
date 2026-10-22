package di

import "darbelis.eu/wshub/ws"

var wsHubInstance *ws.Hub = nil

func GetWsHub() *ws.Hub {
	if wsHubInstance == null {
		wsHubInstance = ws.NewHub("common")
	}

	return wsHubInstance
}
