package lib

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/form"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/skip2/go-qrcode"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const MAINTANCERUN = 30

func writeImagePng(w http.ResponseWriter, buffer []byte) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer)))
	if _, err := w.Write(buffer); err != nil {
		log.Println("Unable to write image.")
	}
}
func getHostnameFromUrl(url string) string {
	reg := regexp.MustCompile("^(?:(?:.*?)?//)?[^/?#;]*")
	return reg.FindString(url)

}
func CheckSign(secret string, value string, sign string) bool {
	md5 := GetMD5Hash(value + secret)
	return strings.ToUpper(md5) == strings.ToUpper(sign)
}

//Base controller
type Controller struct {
}

var repository = Repository{}
var decoder = schema.NewDecoder()
var api = NewApi()

func init() {
	repository.Connect()

	ticker := time.NewTicker(MAINTANCERUN * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				repository.Maintenance()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(response)
}
func (c *Controller) IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.Dir("./public")).ServeHTTP(w, r)
}
func (с *Controller) SqlHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{PARSE_PARAMS_EXEPTION, err.Error()})
		return
	}
	var sqlQuery SqlQuery
	//Check login information
	errDecode := decoder.Decode(&sqlQuery, r.PostForm)
	if errDecode != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{NOT_ENOUGH_PARAMS, errDecode.Error()})
		return
	}
	respondWithJson(w, http.StatusOK, RunMe(sqlQuery.ConString, sqlQuery.Query))
}
func (c *Controller) LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{PARSE_PARAMS_EXEPTION, err.Error()})
		return
	}
	var userLogin UserLogin
	//Check login information
	errDecode := decoder.Decode(&userLogin, r.PostForm)
	if errDecode != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{NOT_ENOUGH_PARAMS, errDecode.Error()})
		return
	}
	isValid, errCheck := repository.CheckUser(userLogin)
	//DB Problem
	if errCheck != nil {
		respondWithJson(w, http.StatusInternalServerError, errCheck)
	}
	//Login information correct
	if isValid == true {
		//GEN token
		expires := time.Now().Add(time.Second * 60 * 60 * 24)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": userLogin.Login,
			"admin":    false,
			"exp":      expires.Unix(),
		})
		tokenString, errorToken := token.SignedString([]byte(GetSecretKey()))
		if errorToken != nil {
			fmt.Println(errorToken)
		}
		cookie := http.Cookie{Name: "token", Value: tokenString, HttpOnly: true, Expires: expires}
		http.SetCookie(w, &cookie)
		json.NewEncoder(w).Encode(JwtToken{Token: tokenString, Expires: expires.Unix()})

		return
	}
	//UnAuthorized
	respondWithJson(w, http.StatusUnauthorized, Exception{UNAUTHORIZED, ""})
}
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
func (c *Controller) LogoutHandler(w http.ResponseWriter, r *http.Request) {

	cookie := http.Cookie{Name: "token", HttpOnly: true, Expires: time.Now()}
	http.SetCookie(w, &cookie)
	json.NewEncoder(w).Encode(Response{OK_RESPONSE, OK_CODE_RESPONSE})
	return
}
func (c *Controller) Terminals(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, repository.Terminals())
}
func (c *Controller) CheckTicketHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{PARSE_PARAMS_EXEPTION, err.Error()})
		return
	}
	var check CheckTiket
	//Check login information
	errDecode := decoder.Decode(&check, r.PostForm)
	if errDecode != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{NOT_ENOUGH_PARAMS, errDecode.Error()})
		return
	}

	respondWithJson(w, OK_CODE_RESPONSE, repository.CheckTicket(check))
}

func (c *Controller) InitInstance(w http.ResponseWriter, r *http.Request) {
	user := User{"demo", "demo", true}
	repository.AddUser(user)
	respondWithJson(w, http.StatusOK, nil)
}
func (c *Controller) TerminalSet(w http.ResponseWriter, r *http.Request) {
	decoder := form.NewDecoder()
	r.ParseForm()
	var terminal Terminal
	err := decoder.Decode(&terminal, r.PostForm)
	if err != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{NOT_ENOUGH_PARAMS, err.Error()})
		return
	}
	repository.SetTerminal(terminal)
	json.NewEncoder(w).Encode(Response{OK_RESPONSE, OK_CODE_RESPONSE})
}
func (c *Controller) TerimalAuthPng(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gate := vars["id"]
	id, err := strconv.Atoi(gate)
	if err != nil {
		respondWithJson(w, http.StatusBadRequest, err)
	}
	terminalAuth := repository.GetAuthTerminalById(int64(id))
	//TODO FIX THAT! ERROR!!!!!
	terminalAuth.Auth.URL = getHostnameFromUrl(r.Referer())
	jsonAuth, errjson := json.Marshal(terminalAuth)
	if err != nil {
		fmt.Printf("Error: %s", errjson)
		return
	}
	png, err := qrcode.Encode(string(jsonAuth), qrcode.Low, 200)
	log.Println(string(jsonAuth))
	if err == nil {
		writeImagePng(w, png)
	}
}
func (c *Controller) AddTerminalHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{PARSE_PARAMS_EXEPTION, err.Error()})
		return
	}
	var terminal Terminal
	//Check login information
	errDecode := decoder.Decode(&terminal, r.PostForm)
	if errDecode != nil {
		log.Println(errDecode)
		respondWithJson(w, http.StatusBadRequest, Exception{NOT_ENOUGH_PARAMS, errDecode.Error()})
		return
	}
	repository.AddTerminal(terminal)
	/*	ex := repository.AddGroup(group)
		if ex != nil {
			respondWithJson(w, http.StatusBadRequest, ex)
		}*/

}
func (c *Controller) GetBuildings(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, OK_CODE_RESPONSE, api.GetBuildings())

}
func (c *Controller) EventInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idin := vars["id"]
	id, _ := strconv.Atoi(idin)
	eventInfo := repository.GetEventInfo(int64(id))
	respondWithJson(w, OK_CODE_RESPONSE, eventInfo)

}
func (c *Controller) Validation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gate := vars["gate"]
	ticket := vars["ticket"]
	sign := vars["sign"]
	gateId, err := strconv.Atoi(gate)
	if err != nil {
		respondWithJson(w, http.StatusBadRequest, err)
	}
	term := repository.GetTerminalById(int64(gateId))
	if term.Secret != "" && CheckSign(term.Secret, ticket, sign) {
		//Correct sign
		resp, _ := repository.ValidateTicket(ticket, term)
		respondWithJson(w, OK_CODE_RESPONSE, resp)
		return
	}
	// Bad sign or gateId

	respondWithJson(w, http.StatusUnauthorized, Exception{Message: "Unauthorized"})

}
func (c *Controller) ValidationRegistration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gate := vars["gate"]
	ticket := vars["ticket"]
	direction := vars["direction"]
	sign := vars["sign"]

	gateId, err := strconv.Atoi(gate)
	if err != nil {
		respondWithJson(w, http.StatusBadRequest, err)
	}
	term := repository.GetTerminalById(int64(gateId))

	if term.Secret != "" && CheckSign(term.Secret, ticket, sign) {
		//Correct sign
		log.Println("1")
		resp, _ := repository.ValidateRegistrateTicket(ticket, term, direction)
		log.Println("2")
		repository.Log(Log{0, ticket, "Result for " + direction + " from gate #" + gate, resp.Result.Code})
		respondWithJson(w, OK_CODE_RESPONSE, resp)
		return
	}
	// Bad sign or gateId
	repository.Log(Log{0, ticket, "Bad sign request from gate #" + gate + " sign - " + sign, http.StatusUnauthorized})
	respondWithJson(w, http.StatusUnauthorized, Exception{Message: "Unauthorized"})

}
func (c *Controller) Registration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gate := vars["gate"]
	ticket := vars["ticket"]
	direction := vars["direction"]
	sign := vars["sign"]

	gateId, err := strconv.Atoi(gate)
	if err != nil {
		respondWithJson(w, http.StatusBadRequest, err)
	}
	term := repository.GetTerminalById(int64(gateId))
	if term.Secret != "" && CheckSign(term.Secret, ticket, sign) {
		//Correct sign
		resp, _ := repository.RegistrateTicket(ticket, term, direction)
		respondWithJson(w, OK_CODE_RESPONSE, resp)
		return
	}
	// Bad sign or gateId
	respondWithJson(w, http.StatusUnauthorized, Exception{Message: "Unauthorized"})

}
func (c *Controller) Groups(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, repository.Groups())
}
func (c *Controller) LogsHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, repository.Logs())
}
func (c *Controller) AddUserHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{PARSE_PARAMS_EXEPTION, err.Error()})
		return
	}
	var user User
	//Check login information
	errDecode := decoder.Decode(&user, r.PostForm)
	if errDecode != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{NOT_ENOUGH_PARAMS, errDecode.Error()})
		return
	}
	ex := repository.AddUser(user)
	if ex != nil {
		respondWithJson(w, http.StatusBadRequest, ex)
	}
}
func (c *Controller) AddGroupHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{PARSE_PARAMS_EXEPTION, err.Error()})
		return
	}
	var group Group
	//Check login information
	errDecode := decoder.Decode(&group, r.PostForm)
	if errDecode != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{NOT_ENOUGH_PARAMS, errDecode.Error()})
		return
	}
	ex := repository.AddGroup(group)
	if ex != nil {
		respondWithJson(w, http.StatusBadRequest, ex)
	}
}
func (c *Controller) AddMasterKeyHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{PARSE_PARAMS_EXEPTION, err.Error()})
		return
	}
	var masterKey MasterKey
	//Check login information
	errDecode := decoder.Decode(&masterKey, r.PostForm)
	if errDecode != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{NOT_ENOUGH_PARAMS, errDecode.Error()})
		return
	}
	ex := repository.AddMasterKey(masterKey)
	if ex != nil {
		respondWithJson(w, http.StatusBadRequest, ex)
	}
}
func (c *Controller) EventsByGroupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idin := vars["id"]
	id, _ := strconv.Atoi(idin)
	log.Println("Read event for " + idin)
	events := repository.GetEventsByGroup(int64(id))
	respondWithJson(w, OK_CODE_RESPONSE, events)
}
func (c *Controller) EventSync(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idin := vars["id"]
	id, _ := strconv.Atoi(idin)
	event, _ := repository.SyncEvent(int64(id))
	repository.Log(Log{0, strconv.FormatInt(event.Id, 10), "Event synced. " + strconv.Itoa(event.TicketsCached) + " tickets cached.", OK_CODE_RESPONSE})
	respondWithJson(w, OK_CODE_RESPONSE, event)
}
func (c *Controller) SetGroupHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{PARSE_PARAMS_EXEPTION, err.Error()})
		return
	}
	var group Group
	//Check login information
	errDecode := decoder.Decode(&group, r.PostForm)
	if errDecode != nil {
		log.Println(errDecode)
		respondWithJson(w, http.StatusBadRequest, Exception{NOT_ENOUGH_PARAMS, errDecode.Error()})
		return
	}
	ex := repository.SetGroup(group)
	if ex != nil {
		respondWithJson(w, http.StatusBadRequest, ex)
	}
	if group.BuildingId != 0 {
		repository.SyncEventsList(group.BuildingId)
	}

}
func (c *Controller) RemoveGroupHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{PARSE_PARAMS_EXEPTION, err.Error()})
		return
	}
	var group Group
	//Check login information
	errDecode := decoder.Decode(&group, r.PostForm)
	if errDecode != nil {
		respondWithJson(w, http.StatusBadRequest, Exception{NOT_ENOUGH_PARAMS, errDecode.Error()})
		return
	}
	ex := repository.RemoveGroup(group)
	if ex != nil {
		respondWithJson(w, http.StatusBadRequest, ex)
	}
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
				respondWithJson(w, http.StatusUnauthorized, Exception{Message: "Not Authorized"})
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
			respondWithJson(w, http.StatusUnauthorized, Exception{Message: "Not Authorized, token not valid"})
		}
	})
}
