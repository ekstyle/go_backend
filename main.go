package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"encoding/json"
	"github.com/ekstyle/go_backend/lib"
	"log"
)


func RootHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(lib.Exception{Message:"Empty response"})
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", RootHandler)
	log.Fatal(http.ListenAndServe(":8080", r))
}