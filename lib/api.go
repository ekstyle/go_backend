package lib

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

const xmlPattern_acs_export_event = `<?xml version="1.0" encoding="utf-8"?>
<request db="%s" module="acs_export_event" format="json">
<param event_id="%d" />
<auth id="api.kassy.ru" />
</request>`
const xmlPattern_table_buildings = `<?xml version="1.0" encoding="utf-8"?>
<request db="ekb" module="table_building" format="json">
    <auth id="api.kassy.ru" />
</request>`

const ENTRY_RESULT_CODE_ACCEPT = 1
const ENTRY_RESULT_CODE_REENTRY = -1
const ENTRY_RESULT_CODE_NOTFOUND = 0

type Api struct {
	Url       string
	Db        string
	SecretKey string
}
type Building struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Address string `json:"address"`

	HallCount string `json:"hall_count"`
}
type TableBuildings struct {
	Db     string `json:"db"`
	Module string `json:"module"`
	Result struct {
		Code     int    `json:"code"`
		Message  string `json:"message"`
		Checksum string `json:"checksum"`
	} `json:"result"`
	Errors  interface{} `json:"errors"`
	Content []Building  `json:"content"`
}

type TicketExport struct {
	TicketID      int    `bson:"ticket_id" json:"ticket_id"`
	EventID       int    `bson:"event_id" json:"-"`
	TicketBarcode string `bson:"ticket_barcode" json:"ticket_barcode"`
	IsEticket     bool   `bson:"is_eticket" json:"is_eticket"`
	TicketSector  string `bson:"ticket_sector" json:"ticket_sector"`
	TicketTitle   string `bson:"ticket_title" json:"ticket_title"`
	TicketPrice   string `bson:"ticket_price" json:"ticket_price"`
	TicketDt      int    `bson:"ticket_dt" json:"ticket_dt"`
	PlaceID       int    `bson:"place_id" json:"place_id"`
	OrderID       int    `bson:"order_id" json:"order_id"`
	CashboxID     int    `bson:"cashbox_id" json:"cashbox_id"`
	CashboxTitle  string `bson:"cashbox_title" json:"cashbox_title"`
	OperatorTitle string `bson:"operator_title" json:"operator_title"`
	CustomerTitle string `bson:"customer_title" json:"customer_title"`
	LastUpdate    int64  `bson:"last_update" json:"-"`
	Source        string `bson:"source" json:"-"`
}
type ACSExportEvent struct {
	Module string `json:"module"`
	Result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"result"`
	Errors  interface{} `json:"errors"`
	Content struct {
		Data struct {
			Event struct {
				EventID    int            `bson:"event_id" json:"event_id"`
				ShowTitle  string         `bson:"show_title" json:"show_title"`
				EventDt    int            `bson:"event_dt" json:"event_dt"`
				ShowID     int            `bson:"show_id" json:"show_id"`
				VenueID    int            `bson:"venue_id" json:"venue_id"`
				VenueTitle string         `bson:"venue_title" json:"venue_title"`
				HallID     int            `bson:"hall_id" json:"hall_id"`
				HallTitle  string         `bson:"hall_title" json:"hall_title"`
				Tickets    []TicketExport `bson:"-" json:"tickets"`
			} `json:"event"`
		} `json:"data"`
	} `json:"content"`
}

func getUrl() string {
	return os.Getenv("API_URL")
}
func getSecretKey() string {
	key := os.Getenv("API_SECRET_KEY")
	return key
}
func GetMD5Hash(text string) string {
	hash := md5.New()
	hash.Write([]byte(text))
	return hex.EncodeToString(hash.Sum(nil))
}

func NewApi() Api {
	return Api{
		Url:       getUrl(),
		Db:        "sandbox",
		SecretKey: getSecretKey(),
	}
}
func (api *Api) Source() string {
	return api.Url + api.Db
}

func (api *Api) Sync() {

}
func (api *Api) GetEventACS(eventid int) ACSExportEvent {
	xml := fmt.Sprintf(xmlPattern_acs_export_event, api.Db, eventid)
	form := url.Values{
		"xml":  {xml},
		"sign": {GetMD5Hash(xml + api.SecretKey)},
	}
	body := bytes.NewBufferString(form.Encode())
	rsp, err := http.Post(api.Url, "application/x-www-form-urlencoded", body)
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()
	body_byte, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		panic(err)
	}
	var acsExportEvent ACSExportEvent
	json.Unmarshal(body_byte, &acsExportEvent)
	log.Println(acsExportEvent.Content.Data.Event.Tickets)
	return acsExportEvent
}
func (api *Api) GetBuildings() []Building {
	form := url.Values{
		"xml":  {xmlPattern_table_buildings},
		"sign": {GetMD5Hash(xmlPattern_table_buildings + api.SecretKey)},
	}
	body := bytes.NewBufferString(form.Encode())
	rsp, err := http.Post(api.Url, "application/x-www-form-urlencoded", body)
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()
	body_byte, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		panic(err)
	}
	var tableBuildings TableBuildings
	json.Unmarshal(body_byte, &tableBuildings)
	return tableBuildings.Content
}
