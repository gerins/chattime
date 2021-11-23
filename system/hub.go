// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package system

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Room name of Hub
	roomName string

	// Channel for sending signal deleting room
	deleteRoom chan string

	// All client connection in this room
	clients map[*Client]bool

	// Inbound messages from the clients.
	Broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub(roomName string, deleteRoomSignal chan string) *Hub {
	return &Hub{
		roomName:   roomName,
		deleteRoom: deleteRoomSignal,
		Broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				if len(h.clients) == 0 {
					h.deleteRoom <- h.roomName
					return
				}
			}

		case message := <-h.Broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:

				default:
					// Channel are full, can indicate client already closed
					close(client.send)
					delete(h.clients, client)

					if len(h.clients) == 0 {
						h.deleteRoom <- h.roomName
						return
					}
				}
			}
		}
	}
}
