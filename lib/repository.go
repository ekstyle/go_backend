package lib

import (
	"gopkg.in/mgo.v2"
	"os"
	"log"
	"gopkg.in/mgo.v2/bson"
	"crypto/md5"
	"encoding/hex"
)

type Repository struct{
	Server   string
	Database string
}
const SALT = "1c2cf9a0a9031262b894fac41f05e656"
const USER_COLLECTION = "users"
var db *mgo.Database
func (r *Repository) Connect() {
	r.Server = os.Getenv("MONGO_URL")
	r.Database = os.Getenv("MONGO_DB")
	session, err := mgo.Dial(r.Server)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to ", r.Server, "with", r.Database, "database.")
	db = session.DB(r.Database)
}
func hashPassword(pass string) string {
	hash := md5.New()
	hash.Write([]byte(pass+SALT))
	return hex.EncodeToString(hash.Sum(nil))
}
func (r *Repository) CheckUser(userLogin UserLogin) (bool, *Exception){
	//result := &User{}
	//db.C("users").Insert(&User{"tester","just test"})
	userCount, errFind := db.C("users").Find(bson.M{"active": true,"login": userLogin.Login,"password": hashPassword(userLogin.Password)}).Count()
	if errFind != nil {
		return false, &Exception {CANT_SELECT_EXEPTION,errFind}
	}
	//Correct user
	if userCount == 1 {
		return true, nil
	}
	//Not found
	return false, nil
}
func (r *Repository) AddUser(user User) (*Exception) {
	//Try to find user
	userCount, errFind := db.C(USER_COLLECTION).Find(bson.M{"login": user.Login}).Count()
	if errFind != nil {
		return &Exception {CANT_SELECT_EXEPTION,errFind}
	}
	if userCount > 0 {
		return &Exception {USER_EXIST_EXEPTION,nil}
	}
	user.Password = hashPassword(user.Password)
	user.Active = true
	errInsert := db.C(USER_COLLECTION).Insert(user)
	if errInsert != nil {
		return &Exception {CANT_INSERT_EXEPTION,errInsert}
	}
	return nil
}