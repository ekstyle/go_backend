package main

import (
	"net/http"
	"github.com/ekstyle/go_backend/lib"
	"log"
	"os"
)
func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Println("PORT environment not set. Use", port)
	}
	return ":" + port
}

func main() {
	r := lib.NewRouter()
	log.Fatal(http.ListenAndServe(GetPort(), r))
}