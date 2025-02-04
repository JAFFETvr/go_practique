package main

import (
	"bytes"
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
var nextID uint = 1

func getUsers(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(users)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	json.NewDecoder(r.Body).Decode(&user)

	mu.Lock()
	user.ID = nextID
	nextID++
	users = append(users, user)
	mu.Unlock()

	json.NewEncoder(w).Encode(user)
	replicateData()
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var updatedUser User
	json.NewDecoder(r.Body).Decode(&updatedUser)

	mu.Lock()
	defer mu.Unlock()
	for i, u := range users {
		if fmt.Sprintf("%d", u.ID) == params["id"] {
			users[i] = updatedUser
			users[i].ID = u.ID
			json.NewEncoder(w).Encode(users[i])
			replicateData()
			return
		}
	}
	http.Error(w, "User not found", http.StatusNotFound)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	mu.Lock()
	defer mu.Unlock()
	for i, u := range users {
		if fmt.Sprintf("%d", u.ID) == params["id"] {
			users = append(users[:i], users[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			replicateData()
			return
		}
	}
	http.Error(w, "User not found", http.StatusNotFound)
}

func replicateData() {
	mu.Lock()
	jsonData, _ := json.Marshal(users)
	mu.Unlock()

	resp, err := http.Post("http://localhost:5001/sync", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error replicating data:", err)
		return
	}
	defer resp.Body.Close()
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/users", getUsers).Methods("GET")
	r.HandleFunc("/users", createUser).Methods("POST")
	r.HandleFunc("/users/{id}", updateUser).Methods("PUT")
	r.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")

	fmt.Println("Server running on port 5000")
	log.Fatal(http.ListenAndServe(":5000", r))
}
