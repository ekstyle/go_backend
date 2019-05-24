package main

import (
	"github.com/ekstyle/go_backend/lib"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Println("PORT environment not set. Use", port)
	}
	return ":" + port
}
func HandlerFs(publicDir string) http.Handler {
	handler := http.FileServer(http.Dir(publicDir))
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_path := req.URL.Path
		// static files
		if strings.Contains(_path, ".") || _path == "/" {
			handler.ServeHTTP(w, req)
			return
		}
		// the all 404 gonna be served as root
		http.ServeFile(w, req, path.Join(publicDir, "/index.html"))
	})
}

func main() {
	r := lib.NewRouter()
	r.PathPrefix("/").Handler(HandlerFs("./public"))

	log.Fatal(http.ListenAndServe(GetPort(), r))
}
