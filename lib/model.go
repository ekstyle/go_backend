package lib

type User struct {
	Login string `bson:"login" schema:"login"`
	Password string `bson:"password" schema:"password"`
	Active bool `bson:"active" schema:"-"`
}
type UserLogin struct {
	Login string `schema:"login"`
	Password string `schema:"password"`
}
type JwtToken struct {
	Token string `json:"token"`
}