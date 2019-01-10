package lib

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
const xmlPattern_page_event_list = `<?xml version="1.0" encoding="utf-8"?>
<request db="%s" module="page_event_list" format="json">
    <filter show_type_id="" show_id="" rollerman_id="" building_id="%d" hall_id="" date_from="%d" date_to="%d" is_recommend="" />
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
				LastUpdate int64          `bson:"last_update" json:"-"`
			} `json:"event"`
		} `json:"data"`
	} `json:"content"`
}

type PageEventList struct {
	Db     string `json:"db"`
	Module string `json:"module"`
	Result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"result"`
	Errors  interface{} `json:"errors"`
	Content struct {
		Subdivision []struct {
			ID          string `json:"id"`
			Db          string `json:"db"`
			City        string `json:"city"`
			Title       string `json:"title"`
			Address     string `json:"address"`
			Phone       string `json:"phone"`
			Tz          string `json:"tz"`
			PriceMarkup string `json:"price_markup"`
			Port        string `json:"port"`
			State       string `json:"state"`
		} `json:"subdivision"`
		Event []struct {
			ID          string      `json:"id"`
			ShowID      string      `json:"show_id"`
			HallID      string      `json:"hall_id"`
			RollermanID string      `json:"rollerman_id"`
			Date        string      `json:"date"`
			PriceMin    string      `json:"price_min"`
			PriceMax    string      `json:"price_max"`
			PriceMarkup interface{} `json:"price_markup"`
			Vacancies   string      `json:"vacancies"`
			TemplateID  interface{} `json:"template_id"`
			IsRecommend string      `json:"is_recommend"`
			IsPrm       string      `json:"is_prm"`
			IsBooking   string      `json:"is_booking"`
			IsSale      string      `json:"is_sale"`
			EventState  string      `json:"event_state"`
			State       string      `json:"state"`
		} `json:"event"`
		ShowType []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			Descr string `json:"descr"`
			Other string `json:"other"`
			Order string `json:"order"`
			State string `json:"state"`
		} `json:"show_type"`
		Show []struct {
			ID          string      `json:"id"`
			TypeID      string      `json:"type_id"`
			Title       string      `json:"title"`
			Announce    string      `json:"announce"`
			Duration    string      `json:"duration"`
			AgeRating   string      `json:"age_rating"`
			Rating      string      `json:"rating"`
			Image       string      `json:"image"`
			VideoID     string      `json:"video_id"`
			PriceMarkup interface{} `json:"price_markup"`
			IsBooking   string      `json:"is_booking"`
			IsSale      string      `json:"is_sale"`
			State       string      `json:"state"`
		} `json:"show"`
		Rollerman []struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			Address string `json:"address"`
			Email   string `json:"email"`
			Phone   string `json:"phone"`
			Inn     string `json:"inn"`
			Okpo    string `json:"okpo"`
			State   string `json:"state"`
		} `json:"rollerman"`
		Hall []struct {
			ID         string `json:"id"`
			BuildingID string `json:"building_id"`
			Title      string `json:"title"`
			Descr      string `json:"descr"`
			TemplateID string `json:"template_id"`
			State      string `json:"state"`
		} `json:"hall"`
		Building []struct {
			ID          string `json:"id"`
			TypeID      string `json:"type_id"`
			CityID      string `json:"city_id"`
			Title       string `json:"title"`
			Descr       string `json:"descr"`
			Address     string `json:"address"`
			Phone       string `json:"phone"`
			URL         string `json:"url"`
			Workhrs     string `json:"workhrs"`
			HallCount   string `json:"hall_count"`
			GeoLat      string `json:"geo_lat"`
			GeoLng      string `json:"geo_lng"`
			PriceMarkup string `json:"price_markup"`
			IsBooking   string `json:"is_booking"`
			IsSale      string `json:"is_sale"`
			State       string `json:"state"`
		} `json:"building"`
	} `json:"content"`
}

func (pg *PageEventList) ShowTitleById(showId string) string {
	for i := range pg.Content.Show {
		if pg.Content.Show[i].ID == showId {
			return pg.Content.Show[i].Title
		}
	}
	return ""
}
func (pg *PageEventList) HallTitleById(hallid string) string {
	for i := range pg.Content.Hall {
		if pg.Content.Hall[i].ID == hallid {
			return pg.Content.Hall[i].Title
		}
	}
	return ""
}

func (pg *PageEventList) ToEvents() Events {

	events := Events{}
	for i := range pg.Content.Event {
		event := Event{}
		event.Title = pg.ShowTitleById(pg.Content.Event[i].ShowID)
		event.Id, _ = strconv.ParseInt(pg.Content.Event[i].ID, 10, 32)
		event.EventDT, _ = strconv.ParseInt(pg.Content.Event[i].Date, 10, 32)
		event.VenueId, _ = strconv.ParseInt(pg.Content.Building[0].ID, 10, 32)
		event.VenueTitle = pg.Content.Building[0].Title
		event.Hall = pg.HallTitleById(pg.Content.Event[i].HallID)
		event.HallId, _ = strconv.ParseInt(pg.Content.Event[i].HallID, 10, 32)
		events.Events = append(events.Events, event)
	}
	return events
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
		Db:        "ekb",
		SecretKey: getSecretKey(),
	}
}
func (api *Api) Source() string {
	return api.Url + api.Db
}

func (api *Api) Sync() {

}
func (api *Api) GetEventACS(eventid int64) ACSExportEvent {
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
	return acsExportEvent
}
func (api *Api) PageEventList(buildingId int64, dtFrom int64, dtTo int64) PageEventList {
	xml := fmt.Sprintf(xmlPattern_page_event_list, api.Db, buildingId, dtFrom, dtTo)
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
	var page PageEventList
	json.Unmarshal(body_byte, &page)
	return page
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
