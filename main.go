// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"chattime/system"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "home.html")
}

func main() {
	flag.Parse()

	http.HandleFunc("/", serveHome)

	// Room storage and management
	roomDatabase := make(map[string]*system.Hub)
	deleteRoomSignal := make(chan string, 10)

	// Delete room listener
	go func(chan string, map[string]*system.Hub) {
		for roomName := range deleteRoomSignal {
			fmt.Printf("Room %v are deleted from database \n", roomName)
			delete(roomDatabase, roomName) // Remove room from roomDatabase
		}
	}(deleteRoomSignal, roomDatabase)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		roomName := r.URL.Query().Get("room")

		// Check room avaibility in map
		if roomHub, exist := roomDatabase[roomName]; exist {
			fmt.Printf("Someone join room %v \n", roomName)
			system.ServeWs(roomHub, w, r)

		} else { // Create new room if not found
			fmt.Printf("Room %v are created \n", roomName)
			roomHub := system.NewHub(roomName, deleteRoomSignal)
			roomDatabase[roomName] = roomHub
			go roomHub.Run()
			system.ServeWs(roomHub, w, r)
		}
	})

	// Special API for broadcasting message to a channel
	http.HandleFunc("/broadcast", func(w http.ResponseWriter, r *http.Request) {
		roomName := r.URL.Query().Get("room")
		message := r.URL.Query().Get("message")

		if roomHub, found := roomDatabase[roomName]; found {
			roomHub.Broadcast <- []byte(message)
		} else {
			http.Error(w, "Bad request, room not found", http.StatusBadRequest)
			return
		}
	})

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
