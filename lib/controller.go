package lib

import (
	"net/http"
	"encoding/json"
	"strings"
	"github.com/dgrijalva/jwt-go"
	"fmt"
	"os"
	"log"
	"github.com/gorilla/context"
	"time"
)
//Base controller
type Controller struct {

}
func GetSecretKey() string {
	key := os.Getenv("SECRET_KEY")
	if key == "" {
		key = "secretKEYmustBEset"
		log.Println("Please set ENV: SECRET_KEY. API not secured!!!")
	}
	return key
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
//Middleware
func AuthenticationMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var clientToken string
		JWTCookie, cookieErr := req.Cookie("token")
		if cookieErr != nil {
			authorizationHeader := req.Header.Get("authorization")
			bearerToken := strings.Split(authorizationHeader, " ")
			if len(bearerToken) != 2 {
				//http.Redirect(w, req, "/get-token", 301)
				respondWithJson(w, http.StatusUnauthorized, Exception{Message:"Not Authorized"})
				return
			}
			clientToken = bearerToken[1]
		} else {
			clientToken = JWTCookie.Value
		}
		token, error := jwt.Parse(clientToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("There was an error")
			}
			return []byte(GetSecretKey()), nil
		})
		if error != nil {
			json.NewEncoder(w).Encode(Exception{Message: error.Error()})
			return
		}
		if token.Valid {
			log.Println("TOKEN WAS VALID")
			context.Set(req, "decoded", token.Claims)
			next(w, req)
		} else {
			respondWithJson(w, http.StatusUnauthorized, Exception{Message:"Not Authorized, token not valid"})
		}
	})
}
//GEN token
func (c *Controller) GetToken(w http.ResponseWriter, req *http.Request) {

	expires := time.Now().Add(time.Second * 60)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "test",
		"admin": false,
		"exp":      expires.Unix(),
	})
	tokenString, error := token.SignedString([]byte(GetSecretKey()))
	if error != nil {
		fmt.Println(error)
	}
	cookie := http.Cookie{Name: "token", Value: tokenString, HttpOnly: true, Expires: expires,}
	http.SetCookie(w, &cookie)
	json.NewEncoder(w).Encode(JwtToken{Token: tokenString})
}