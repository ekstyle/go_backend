package lib

type Response struct {
	Result string `json:"result"`
	Code   int    `json:"code"`
}

const OK_RESPONSE = "OK"

const OK_CODE_RESPONSE = 200
