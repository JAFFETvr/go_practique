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

var (
	users       []User
	usersMutex  sync.Mutex
	subscribers []chan []User 
)

func getUsers(w http.ResponseWriter, r *http.Request) {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	json.NewEncoder(w).Encode(users)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	usersMutex.Lock()
	user.ID = uint(len(users) + 1) 
	users = append(users, user)
	usersMutex.Unlock()

	notifySubscribers()
	replicateData()
	json.NewEncoder(w).Encode(user)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
    log.Println(" Intentando actualizar usuario...")
    params := mux.Vars(r)
    var idToUpdate uint

    if _, err := fmt.Sscanf(params["id"], "%d", &idToUpdate); err != nil {
        log.Println(" Error al leer el ID:", err)
        http.Error(w, "ID inválido", http.StatusBadRequest)
        return
    }

    usersMutex.Lock()
    defer usersMutex.Unlock()

    log.Println("Buscando usuario con ID:", idToUpdate)

    for i, user := range users {
        if user.ID == idToUpdate {
            log.Println(" Usuario encontrado:", user)

            var updatedUser User
            if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
                log.Println(" Error al decodificar JSON:", err)
                http.Error(w, "Datos inválidos", http.StatusBadRequest)
                return
            }

            updatedUser.ID = idToUpdate // Mantiene el mismo ID
            users[i] = updatedUser

            log.Println(" Usuario actualizado:", users[i])

            notifySubscribers()
            replicateData()
            json.NewEncoder(w).Encode(users[i])
            return
        }
    }

    log.Println(" Usuario no encontrado")
    http.Error(w, "Usuario no encontrado", http.StatusNotFound)
}


func deleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var idToDelete uint
	fmt.Sscanf(params["id"], "%d", &idToDelete)

	usersMutex.Lock()
	for i, user := range users {
		if user.ID == idToDelete {
			users = append(users[:i], users[i+1:]...)
			usersMutex.Unlock()
			notifySubscribers()
			replicateData()
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	usersMutex.Unlock()
	http.Error(w, "User not found", http.StatusNotFound)
}

func replicateData() {
    usersMutex.Lock()
    jsonData, err := json.Marshal(users)
    usersMutex.Unlock()

    if err != nil {
        log.Println(" Error al serializar datos:", err)
        return
    }

    log.Println("Enviando datos al servidor réplica...")

    resp, err := http.Post("http://localhost:5001/sync", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        log.Println("Error al replicar datos:", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Println(" Respuesta inesperada del servidor réplica:", resp.Status)
    } else {
        log.Println("Datos replicados correctamente")
    }
}


func longPolling(w http.ResponseWriter, r *http.Request) {
	updateChan := make(chan []User)

	usersMutex.Lock()
	subscribers = append(subscribers, updateChan)
	usersMutex.Unlock()

	users := <-updateChan 
	json.NewEncoder(w).Encode(users)
}

func notifySubscribers() {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	for _, ch := range subscribers {
		ch <- users
	}
	subscribers = nil
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/users", getUsers).Methods("GET")
	r.HandleFunc("/users", createUser).Methods("POST")
	r.HandleFunc("/users/{id}", updateUser).Methods("PUT")
	r.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")
	r.HandleFunc("/longpolling", longPolling).Methods("GET")

	fmt.Println("Servidor Principal en puerto 5000")
	log.Fatal(http.ListenAndServe(":5000", r))
}
