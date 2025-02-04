package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type User struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	User string `json:"user"`
}

var users []User // Almacenamiento local en la réplica

func syncData(w http.ResponseWriter, r *http.Request) {
	var newUsers []User
	json.NewDecoder(r.Body).Decode(&newUsers)
	users = newUsers
	fmt.Println(" Réplica actualizada:", users)
	w.WriteHeader(http.StatusOK)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(users)
}

func shortPolling() {
	for {
		resp, err := http.Get("http://localhost:5000/users")
		if err == nil {
			var newUsers []User
			json.NewDecoder(resp.Body).Decode(&newUsers)
			users = newUsers
		} else {
			log.Println("Error en Short Polling:", err)
		}
		time.Sleep(5 * time.Second)
	}
}

func longPolling() {
	for {
		resp, err := http.Get("http://localhost:5000/longpolling")
		if err == nil {
			var newUsers []User
			json.NewDecoder(resp.Body).Decode(&newUsers)
			users = newUsers
			fmt.Println("📡 Long Polling: Datos actualizados")
		} else {
			log.Println("Error en Long Polling:", err)
		}
	}
}

func main() {
	go shortPolling() 
	go longPolling()  

	r := mux.NewRouter()
	r.HandleFunc("/users", getUsers).Methods("GET") // Solo lectura
	r.HandleFunc("/sync", syncData).Methods("POST") 

	fmt.Println("Servidor Réplica en puerto 5001")
	log.Fatal(http.ListenAndServe(":5001", r))
}
