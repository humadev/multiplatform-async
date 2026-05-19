package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

// --- CRUD Model and Memory Storage ---

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

var (
	items  = make(map[int]Item)
	nextID = 1
	mu     sync.Mutex
)

// --- CRUD Handlers ---

func handleItems(w http.ResponseWriter, r *http.Request) {
	// Setup CORS for local testing if needed
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		getItems(w, r)
	case http.MethodPost:
		createItem(w, r)
	case http.MethodPut:
		updateItem(w, r)
	case http.MethodDelete:
		deleteItem(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getItems(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var itemList []Item
	for _, item := range items {
		itemList = append(itemList, item)
	}

	json.NewEncoder(w).Encode(itemList)
}

func createItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	item.ID = nextID
	nextID++
	items[item.ID] = item
	mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func updateItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || idStr == "" {
		http.Error(w, "invalid or missing id query parameter", http.StatusBadRequest)
		return
	}

	var updatedItem Item
	if err := json.NewDecoder(r.Body).Decode(&updatedItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := items[id]; !exists {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}

	updatedItem.ID = id
	items[id] = updatedItem
	json.NewEncoder(w).Encode(updatedItem)
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || idStr == "" {
		http.Error(w, "invalid or missing id query parameter", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := items[id]; !exists {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}

	delete(items, id)
	w.WriteHeader(http.StatusNoContent)
}

// --- WebSocket Chat Storage ---

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demonstration
	},
}

type Client struct {
	conn *websocket.Conn
}

var (
	clients   = make(map[*Client]bool)
	broadcast = make(chan []byte)
	chatMu    sync.Mutex
)

// --- WebSocket Handlers ---

func handleChatConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer ws.Close()

	client := &Client{conn: ws}

	chatMu.Lock()
	clients[client] = true
	chatMu.Unlock()

	log.Println("New client connected to chat")

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Error reading message or client disconnected:", err)
			chatMu.Lock()
			delete(clients, client)
			chatMu.Unlock()
			break
		}
		// Send the message to the broadcast channel
		broadcast <- msg
	}
}

func handleChatMessages() {
	for {
		// Grab next message from broadcast channel
		msg := <-broadcast
		
		chatMu.Lock()
		for client := range clients {
			err := client.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("Error writing message to client:", err)
				client.conn.Close()
				delete(clients, client)
			}
		}
		chatMu.Unlock()
	}
}

func main() {
	// Register HTTP handlers
	http.HandleFunc("/api/items", handleItems)
	
	// Register WebSocket handler
	http.HandleFunc("/ws/chat", handleChatConnections)

	// Start message broadcasting goroutine
	go handleChatMessages()

	port := ":8080"
	fmt.Printf("Server started on http://localhost%s\n", port)
	fmt.Printf("CRUD endpoint: http://localhost%s/api/items\n", port)
	fmt.Printf("Chat endpoint: ws://localhost%s/ws/chat\n", port)
	
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
