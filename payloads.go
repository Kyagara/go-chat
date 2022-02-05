package main

import (
	"encoding/json"
	"log"
)

type Message struct {
	// Tipo da mensagem, 'event' ou 'command'
	Type string `json:"type"`
	// Tópico da mensagem, 'room', 'game'
	Topic string `json:"topic"`
	// Nome do evento ou comando
	Name string `json:"name"`
	// Payload da mensagem, map[string]interface{}
	Data JSON `json:"data"`
}

// Objeto json com key do tipo string
type JSON map[string]interface{}

var (
	roomNotFound   = Message{Type: "error", Topic: "room", Name: "not_found", Data: JSON{"message": "Sala não encontrada"}}
	clientNotFound = Message{Type: "error", Topic: "client", Name: "not_found", Data: JSON{"message": "Cliente não encontrado"}}
	cmdBadRequest  = Message{Type: "error", Topic: "command", Name: "bad_request", Data: JSON{"message": "Comando inválido"}}
	internalError  = Message{Type: "error", Topic: "command", Name: "internal_error", Data: JSON{"message": "Ocorreu um erro interno"}}
	ack            = Message{Type: "event", Topic: "client", Name: "ack", Data: JSON{"ok": true}}
)

// Valida e retorna uma mensagem
func ParseMessage(msg []byte) *Message {
	var message *Message

	err := json.Unmarshal(msg, &message)

	if err != nil {
		log.Println("Erro ao validar uma mensagem:", err)
	}

	return message
}
