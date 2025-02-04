package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type User struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	User string `json:"user"`
}

var users []User
var mu sync.Mutex

func syncUsers(w http.ResponseWriter, r *http.Request) {
	var receivedUsers []User
	json.NewDecoder(r.Body).Decode(&receivedUsers)

	mu.Lock()
	users = receivedUsers
	mu.Unlock()

	fmt.Println("Data replicated successfully")
	w.WriteHeader(http.StatusOK)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(users)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/sync", syncUsers).Methods("POST")
	r.HandleFunc("/users", getUsers).Methods("GET")

	fmt.Println("Replica server running on port 5001")
	log.Fatal(http.ListenAndServe(":5001", r))
}
