package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Message struct {
	Sender string `json:"Sender"`
	Text   string `json:"Text"`
}

var (
	mu       sync.Mutex
	messages []Message
)

func handleChat(w http.ResponseWriter, r *http.Request) {
	// السماح بالاتصال من أي مصدر (CORS)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// استقبال رسالة جديدة (POST)
	if r.Method == "POST" {
		var msg Message
		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mu.Lock()
		messages = append(messages, msg)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
		return
	}

	// جلب كل الرسائل (GET)
	if r.Method == "GET" {
		mu.Lock()
		defer mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
		return
	}
}

func main() {
	http.HandleFunc("/messages", handleChat)

	port := ":1234"
	fmt.Printf("HTTP Chat Server running on port %s...\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}
