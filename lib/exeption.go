package lib

type Exception struct {
	Message string `json:"message"`
	err error
}
const CANT_INSERT_EXEPTION = "Can`t insert in database"
const CANT_SELECT_EXEPTION = "Can`t select from database"
const USER_EXIST_EXEPTION = "Can`t add, user already exists"
const PARSE_PARAMS_EXEPTION = "Can`t parse params"
const NOT_ENOUGH_PARAMS = "Not enouth params"
const UNAUTHORIZED = "Unauthorized access "