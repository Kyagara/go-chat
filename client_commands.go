package main

import (
	"encoding/json"
	"time"

	"github.com/lucsky/cuid"
)

func ClientCommandHandler(c *Client, cmd *Message) Message {
	data, _ := json.Marshal(cmd.Data)

	var payload JSON

	json.Unmarshal(data, &payload)

	switch cmd.Name {
	case "direct_message": // Enviar mensagem para um cliente
		clientId, ok := payload["client"].(string)

		if !ok || clientId == "" {
			return cmdBadRequest
		}

		message, ok := payload["msg"].(string)

		if !ok || message == "" {
			return cmdBadRequest
		}

		client, ok := c.Hub.Clients[clientId]

		if ok {
			data := JSON{"id": cuid.New(), "client": clientId, "ts": time.Now().Unix(), "msg": message}

			payload, _ := json.Marshal(&Message{
				Type:  "event",
				Topic: "client",
				Name:  "direct_message",
				Data:  data,
			})

			client.Send <- payload

			return ack
		}

		return clientNotFound

	default:
		return Message{Type: "error", Topic: "command", Name: "not_implemented", Data: JSON{"type": cmd.Type, "topic": cmd.Topic, "name": cmd.Name, "data": cmd.Data}}
	}
}
