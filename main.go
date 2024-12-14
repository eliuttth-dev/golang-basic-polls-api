package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Poll struct {
	ID       int            `json:"id"`
	Question string         `json:"question"`
	Options  map[string]int `json:"options"`
}

var (
	polls   = []Poll{}
	nextID  = 1
	mutex   = &sync.Mutex{}
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/polls", getPolls).Methods("GET")
	router.HandleFunc("/polls/{id:[0-9]+}", getPollByID).Methods("GET")
	router.HandleFunc("/polls", createPoll).Methods("POST")
	router.HandleFunc("/polls/{id:[0-9]+}/vote", votePoll).Methods("POST")
	router.HandleFunc("/polls/{id:[0-9]+}", deletePoll).Methods("DELETE")

	log.Println("Polls API running on :3000")
	log.Fatal(http.ListenAndServe(":3000", router))
}

func getPolls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(polls)
}

func getPollByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	for _, poll := range polls {
		if poll.ID == id {
			json.NewEncoder(w).Encode(poll)
			return
		}
	}
	http.Error(w, "Poll not found", http.StatusNotFound)
}

func createPoll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newPoll Poll
	if err := json.NewDecoder(r.Body).Decode(&newPoll); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	newPoll.ID = nextID
	nextID++
	if newPoll.Options == nil {
		newPoll.Options = make(map[string]int)
	}
	polls = append(polls, newPoll)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPoll)
}

func votePoll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	var vote struct {
		Option string `json:"option"`
	}
	if err := json.NewDecoder(r.Body).Decode(&vote); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	for i, poll := range polls {
		if poll.ID == id {
			if _, exists := poll.Options[vote.Option]; !exists {
				http.Error(w, "Option not found in poll", http.StatusBadRequest)
				return
			}

			polls[i].Options[vote.Option]++
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "Poll not found", http.StatusNotFound)
}

func deletePoll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	mutex.Lock()
	defer mutex.Unlock()

	for i, poll := range polls {
		if poll.ID == id {
			polls = append(polls[:i], polls[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "Poll not found", http.StatusNotFound)
}

