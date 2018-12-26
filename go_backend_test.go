package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ekstyle/go_backend/lib"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

const GATE = "1"
const BARCODEFORTEST = "3620897234348"
const SECRETKEY = "5d9f2f8931434f346faf8a17be68f0d1"
const MASTERKEY = "all open 001"

const FAILSKDRESPONSE = -100
const OK = 1
const FAIL = -1
const NOTFOUND = 0

func RandomInt(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}
func RandomStr(len int) string {
	var result bytes.Buffer
	var temp string
	for i := 0; i < len; {
		if string(RandomInt(65, 90)) != temp {
			temp = string(RandomInt(65, 90))
			result.WriteString(temp)
			i++
		}
	}
	return result.String()
}

func TesterGET(urltest string) ([]byte, error) {
	var buf []byte
	url := fmt.Sprintf("http://localhost%s/%s", GetPort(), urltest)
	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {

		return buf, err
	}
	if resp.StatusCode != http.StatusOK {
		return buf, fmt.Errorf("Error GET %s: %s", url, resp.Status)
	}
	buf, _ = ioutil.ReadAll(resp.Body)
	return buf, nil

}

type SKDRequest struct {
	gateid    string
	direction string
	barcode   string
	secretKey string
}

func GetSKDResponseCodeValidateRegistrate(rep interface{}) (int64, error) {
	req := rep.(SKDRequest)
	urlentry := fmt.Sprintf("/validation/%s/%s/%s?sign=%s", req.gateid, req.direction, req.barcode, lib.GetMD5Hash(req.barcode+req.secretKey))
	resp, err := TesterGET(urlentry)
	if err != nil {
		return FAILSKDRESPONSE, err
	}
	skdResp := lib.SKDResponse{}
	json.Unmarshal(resp, &skdResp)
	return skdResp.Result.Code, nil
}

func TestGetPort(t *testing.T) {
	if GetPort() == "" {
		t.Error("Cant Get port")
	}
}

func TestGetBuildings(t *testing.T) {
	resp, err := TesterGET("/buildings")
	if err != nil {
		t.Error("Error", err, resp)
	}
	t.Log("Building: ", string(resp))
}

type fibTestFunc func(interface{}) (int64, error)

func TestValidationRegistration(t *testing.T) {
	var fibTests = []struct {
		testData interface{}
		fn       fibTestFunc
		expected int64
	}{
		{SKDRequest{GATE, "entry", RandomStr(10), SECRETKEY}, GetSKDResponseCodeValidateRegistrate, NOTFOUND},             //#1)Not found
		{SKDRequest{GATE, "entry", BARCODEFORTEST, SECRETKEY}, GetSKDResponseCodeValidateRegistrate, OK},                  //#2)Correct
		{SKDRequest{GATE, "entry", BARCODEFORTEST, SECRETKEY}, GetSKDResponseCodeValidateRegistrate, FAIL},                //#3)Reentry
		{SKDRequest{GATE, "exit", BARCODEFORTEST, SECRETKEY}, GetSKDResponseCodeValidateRegistrate, FAIL},                 //#4)Exit for no entry
		{SKDRequest{GATE, "entry", BARCODEFORTEST, RandomStr(32)}, GetSKDResponseCodeValidateRegistrate, FAILSKDRESPONSE}, //#5)Not correct sign
		{SKDRequest{GATE, "entry", MASTERKEY, SECRETKEY}, GetSKDResponseCodeValidateRegistrate, OK},                       //#6)MasterKey
	}
	for idx, tt := range fibTests {
		actual, _ := tt.fn(tt.testData)
		if actual != tt.expected {
			t.Errorf("(#%d) FibTest fail %s for(%s): expected %d, actual %d", idx+1, tt.testData.(SKDRequest).direction, tt.testData.(SKDRequest).barcode, tt.expected, actual)
		}
	}
}
