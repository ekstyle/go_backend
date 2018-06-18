package lib

import (
	"net/http"
	"encoding/json"
)
//Base controller
type Controller struct {

}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
func (c *Controller) Index(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, Exception{Message:"Empty"})
}