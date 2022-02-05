package main

import (
	"encoding/json"
	"time"

	"github.com/lucsky/cuid"
)

func RoomCommandHandler(c *Client, cmd *Message) Message {
	data, _ := json.Marshal(cmd.Data)

	var payload JSON

	json.Unmarshal(data, &payload)

	switch cmd.Name {
	case "create_room": // Criar uma sala
		// Criando sala
		room := NewRoom(c)

		// Iniciando goroutine da sala
		go room.run()

		// Adicionando a sala no hub
		c.Hub.AddRoom <- room
		room.AddClient <- c

		return Message{Type: "event", Topic: "room", Name: "new_room", Data: JSON{"room": room.Id}}

	case "delete_room": // Deletar uma sala
		roomId, ok := payload["room"].(string)

		if !ok || roomId == "" {
			return cmdBadRequest
		}

		room, ok := c.Hub.Rooms[roomId]

		if ok && room.Admin.Id == c.Id {
			c.Hub.RemoveRoom <- room
			return ack
		}

		return roomNotFound

	case "join_room": // Entrar em uma sala
		roomId, ok := payload["room"].(string)

		if !ok || roomId == "" {
			return cmdBadRequest
		}

		room, ok := c.Hub.Rooms[roomId]

		if ok && c.Rooms[roomId] == nil {
			room.AddClient <- c
			return ack
		}

		return roomNotFound

	case "leave_room": // Sair de uma sala
		roomId, ok := payload["room"].(string)

		if !ok || roomId == "" {
			return cmdBadRequest
		}

		room, ok := c.Rooms[roomId]

		if ok && c.Rooms[roomId] != nil {
			room.RemoveClient <- c
			return ack
		}

		return roomNotFound

	case "room_info": // Obter informações sobre uma sala
		roomId, ok := payload["room"].(string)

		if !ok || roomId == "" {
			return cmdBadRequest
		}

		room, ok := c.Rooms[roomId]

		if ok {
			data := JSON{"id": room.Id, "clients": len(room.Clients)}

			return Message{Type: "event", Topic: "room", Name: "info", Data: data}
		}

		return roomNotFound

	case "broadcast": // Enviar mensagem para uma sala
		roomId, ok := payload["room"].(string)

		if !ok || roomId == "" {
			return cmdBadRequest
		}

		message, ok := payload["msg"].(string)

		if !ok || message == "" {
			return cmdBadRequest
		}

		if c.Rooms[roomId] != nil {
			json, _ := json.Marshal(JSON{"id": cuid.New(), "room": roomId, "ts": time.Now().Unix(), "msg": message})

			err := rdb.Publish(ctx, roomId, json).Err()

			if err != nil {
				panic(err)
			}

			return ack
		}

		return roomNotFound

	default:
		return Message{Type: "error", Topic: "command", Name: "not_implemented", Data: JSON{"type": cmd.Type, "topic": cmd.Topic, "name": cmd.Name, "data": cmd.Data}}
	}
}
