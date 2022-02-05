package main

import (
	"encoding/json"
	"log"

	"github.com/lucsky/cuid"
)

// Sala de clientes conectada ao hub
type Room struct {
	// Id da sala, um CUID
	Id string
	// Administrador da sala
	Admin *Client
	// Clientes na sala
	Clients map[string]*Client
	// Adicionar um cliente na sala
	AddClient chan *Client
	// Remover um cliente da sala
	RemoveClient chan *Client
	// Enviar mensagem para todos os clientes de uma sala
	Broadcast chan *Message
}

func NewRoom(c *Client) *Room {
	return &Room{
		Id:           cuid.New(),
		Admin:        c,
		Clients:      make(map[string]*Client),
		AddClient:    make(chan *Client),
		RemoveClient: make(chan *Client),
		Broadcast:    make(chan *Message, 256),
	}
}

func (r *Room) run() {
	go r.subscribe()

	for {
		select {
		case client := <-r.AddClient:
			r.registerClient(client)

		case client := <-r.RemoveClient:
			r.unregisterClient(client)

		case message := <-r.Broadcast:
			r.broadcastToClients(message)
		}
	}
}

func (r *Room) registerClient(c *Client) {
	// Adicionando o cliente na sala
	r.Clients[c.Id] = c
	c.Rooms[r.Id] = r

	log.Printf("Cliente '%v' registrado na sala '%v'", c.Id, r.Id)

	exists, _ := rdb.SIsMember(ctx, c.Id, r.Id).Result()

	if !exists {
		rdb.SAdd(ctx, c.Id, r.Id)

		// Broadcast anunciando que um cliente novo entrou na sala
		r.Broadcast <- &Message{Type: "event", Topic: "room", Name: "client_joined", Data: JSON{"room": r.Id, "client": c.Id}}
	}
}

func (r *Room) unregisterClient(c *Client) {
	if r.Clients[c.Id] != nil {
		// Removendo cliente
		delete(r.Clients, c.Id)
		delete(c.Rooms, r.Id)

		log.Printf("Cliente '%v' removido da sala '%v'", c.Id, r.Id)

		exists, _ := rdb.SIsMember(ctx, c.Id, r.Id).Result()

		if exists {
			rdb.SRem(ctx, c.Id, r.Id)

			r.Broadcast <- &Message{Type: "event", Topic: "room", Name: "client_left", Data: JSON{"room": r.Id, "client": c.Id}}
		}

		// Caso a sala não tenha mais nenhum cliente
		if len(r.Clients) == 0 {
			c.Hub.RemoveRoom <- r
		}
	}
}

// Iniciando a goroutine de PubSub da sala
func (r *Room) subscribe() {
	// Subscribe em um canal chamado messages
	subscription := rdb.Subscribe(ctx, r.Id)

	// Criando um go channel com o channel da redis
	channel := subscription.Channel()

	for message := range channel {
		var data JSON
		json.Unmarshal([]byte(message.Payload), &data)

		r.Broadcast <- &Message{Type: "event", Topic: "room", Name: "broadcast", Data: data}
	}
}

// Broadcast em uma mensagem para todos os clientes conectados na sala
func (r *Room) broadcastToClients(msg *Message) {
	data, err := json.Marshal(msg)

	if err != nil {
		log.Printf("Erro '%v' ao validar mensagem de Broadcast: %v", msg, err)
		return
	}

	// Enviando a mensagem para todos os clientes
	for _, client := range r.Clients {
		client.Send <- data
	}
}
