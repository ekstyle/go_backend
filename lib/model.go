package lib

type User struct {
	Login string `bson:"login"`
	Password string `bson:"password"`
	Active bool `bson:"active"`
}
type UserLogin struct {
	Login string `schema:"login"`
	Password string `schema:"password"`
}
type JwtToken struct {
	Token string `json:"token"`
}