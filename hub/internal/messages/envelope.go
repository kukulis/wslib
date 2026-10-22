package messages

import "darbelis.eu/wshub/message_common"
import "encoding/json"

type Envelope struct {
	Id   string                     `json:"id"`
	Type message_common.MessageType `json:"type"`
	Data json.RawMessage            `json:"data"`
}
