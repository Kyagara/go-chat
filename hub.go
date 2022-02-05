package main

import (
	"log"
)

// Estrutura que representa uma sala principal onde todos os clientes conectados no websockets são encontrados
type Hub struct {
	// Clientes registrados no hub
	Clients map[string]*Client
	// Salas registradas no hub
	Rooms map[string]*Room
	// Adicionar um cliente no hub
	AddClient chan *Client
	// Remover um cliente do hub
	RemoveClient chan *Client
	// Adicionar uma sala no hub
	AddRoom chan *Room
	// Remover uma sala do hub
	RemoveRoom chan *Room
}

func NewHub() *Hub {
	return &Hub{
		Clients:      make(map[string]*Client),
		Rooms:        make(map[string]*Room),
		AddClient:    make(chan *Client),
		RemoveClient: make(chan *Client),
		AddRoom:      make(chan *Room),
		RemoveRoom:   make(chan *Room),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.AddClient:
			h.registerClient(client)

		case client := <-h.RemoveClient:
			h.unregisterClient(client)

		case room := <-h.AddRoom:
			h.registerRoom(room)

		case room := <-h.RemoveRoom:
			h.unregisterRoom(room)
		}
	}
}

func (h *Hub) registerClient(c *Client) {
	// Adicionando cliente no mapa de clientes do hub
	h.Clients[c.Id] = c

	log.Printf("Cliente '%v' registrado no hub", c.Id)

	rooms, err := rdb.SMembers(ctx, c.Id).Result()

	if err != nil {
		panic(err)
	}

	for _, roomId := range rooms {
		room, ok := h.Rooms[roomId]

		if ok {
			room.AddClient <- c
		}
	}
}

func (h *Hub) unregisterClient(c *Client) {
	if h.Clients[c.Id] != nil {
		// Fechando o canal de send do cliente e deleta o cliente do mapa de clientes no hub
		close(c.Send)
		delete(h.Clients, c.Id)

		log.Printf("Cliente '%v' removido do hub", c.Id)
	}
}

func (h *Hub) registerRoom(r *Room) {
	// Adicionando a sala no mapa de salas do hub
	h.Rooms[r.Id] = r

	log.Printf("Sala '%v' criada", r.Id)
}

func (h *Hub) unregisterRoom(r *Room) {
	if h.Rooms[r.Id] != nil {
		// Se a sala possuir clientes, remover eles
		for _, c := range r.Clients {
			r.RemoveClient <- c
		}

		// Fechando o canal de broadcast da sala e deleta a sala do mapa de salas no hub
		close(r.Broadcast)
		delete(h.Rooms, r.Id)

		log.Printf("Sala '%v' removida", r.Id)
	}
}
