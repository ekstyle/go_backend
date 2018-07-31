package lib

import (
	"github.com/gorilla/mux"
	"net/http"
)

var controller = &Controller{}

type Route struct {
	Name        string
	Method      string
	Queries     string
	Hadle       string
	Pattern     string
	HandlerFunc http.HandlerFunc
}
type Routes []Route

var routes = Routes{

	Route{
		"Login",
		"POST",
		"", "",
		"/login", controller.LoginHandler,
	},
	Route{
		"Logout",
		"GET",
		"", "",
		"/logout", AuthenticationMiddleware(controller.LogoutHandler),
	},
	Route{
		"AddUser",
		"POST",
		"", "",
		"/add_user", AuthenticationMiddleware(controller.AddUserHandler),
	},
	Route{
		"Terminals",
		"GET",
		"", "",
		"/terminals/{group}/", controller.Terminals,
	},
	Route{
		"TerminalsAuth",
		"GET",
		"", "",
		"/terminal/{id}/auth.png", controller.TerimalAuthPng,
	},
	Route{
		"Groups",
		"GET",
		"", "",
		"/groups", AuthenticationMiddleware(controller.Groups),
	},
	Route{
		"AddGroup",
		"POST",
		"", "",
		"/add_group", AuthenticationMiddleware(controller.AddGroupHandler),
	},
	Route{
		"SetGroup",
		"POST",
		"", "",
		"/set_group", AuthenticationMiddleware(controller.SetGroupHandler),
	},
	Route{
		"RemoveGroup",
		"POST",
		"", "",
		"/remove_group", AuthenticationMiddleware(controller.RemoveGroupHandler),
	},
	Route{
		"Buildings",
		"get",
		"", "",
		"/buildings", controller.GetBuildings,
	},
	Route{
		"SQL",
		"POST",
		"", "",
		"/sql", AuthenticationMiddleware(controller.SqlHandler),
	},
	Route{
		"Validation",
		"GET",
		"sign", "{sign}",
		"/validation/{gate}/{ticket}",
		controller.Validation,
	},
	Route{
		"ValidationRegistration",
		"GET",
		"sign", "{sign}",
		"/validation/{gate}/{direction:entry|exit}/{ticket}",
		controller.ValidationRegistration,
	},
	Route{
		"Registration",
		"GET",
		"sign", "{sign}",
		"/registration/{gate}/{direction:entry|exit}/{ticket}",
		controller.Registration,
	},
}

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		if route.Queries == "" {
			router.
				Methods(route.Method).
				Path(route.Pattern).
				Name(route.Name).
				Handler(handler)
		} else {
			router.
				Methods(route.Method).
				Path(route.Pattern).
				Name(route.Name).
				Handler(handler).
				Queries(route.Queries, route.Hadle)
		}
	}
	return router
}
