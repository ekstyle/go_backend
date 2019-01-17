package lib

import (
	"time"
)

const ( // iota is reset to 0
	Entry_entry = iota // c0 == 0
	Entry_exit  = iota // c1 == 1
)

const TICKET_LOCK_TIME_FOR_REENTRY = 10

type MasterKey struct {
	Barcode string `json:"barcode" bson:"barcode"`
}
type MasterKeys struct {
	keys []MasterKey
}

func (r *MasterKeys) is(barode string) bool {
	for i := 0; i < len(r.keys); i++ {
		//delete old
		if r.keys[i].Barcode == barode {
			return true
		}
	}
	return false
}
func (r *MasterKeys) add(barode string) {
	if !r.is(barode) {
		r.keys = append(r.keys, MasterKey{barode})
	}
}

type User struct {
	Login    string `bson:"login" schema:"login,required"`
	Password string `bson:"password" schema:"password,required"`
	Active   bool   `bson:"active" schema:"-"`
}
type AuthStruct struct {
	Auth struct {
		URL       string `json:"url" bson:"-"`
		ID        int64  `json:"id" bson:"id"`
		Title     string `json:"title" bson:"name"`
		SecretKey string `json:"secret_key" bson:"secret_key"`
	} `json:"auth"`
	Sign string `json:"sign"`
}
type UserLogin struct {
	Login    string `schema:"login,required"`
	Password string `schema:"password,required"`
}
type CheckTiket struct {
	Barcode string `json:"barcode" bson:"barcode"`
}
type CheckResult struct {
	Event  Event  `json:"event"`
	Ticket Ticket `json:"ticket"`
}
type Sign struct {
	Sign string `schema:"sign,required"`
}
type SqlQuery struct {
	ConString string `schema:"constring,required"`
	Query     string `schema:"query,required"`
}
type JwtToken struct {
	Token   string `json:"-"`
	Expires int64  `json:"exp"`
}
type Terminal struct {
	Name   string  `json:"name" bson:"name" form:"name"`
	Id     int64   `json:"id" bson:"id" schema:"id" form:"id"`
	Secret string  `json:"-" bson:"secret_key,omitempty" schema:"-"`
	Groups []int64 `json:"groups" bson:"groups" schema:"-" form:"groups"`
}

type Ticket struct {
	TicketId      int64  `json:"id,omitempty" bson:"ticket_id"`
	EventId       int64  `json:"-" bson:"event_id"`
	TicketBarcode string `json:"barcode" bson:"ticket_barcode"`
	IsEticket     bool   `json:"is_eticket,omitempty" bson:"is_eticket"`
	TicketSector  string `json:"sector,omitempty" bson:"ticket_sector"`
	TicketTitle   string `json:"title,omitempty" bson:"ticket_title"`
	TicketPrice   string `json:"price,omitempty" bson:"ticket_price"`
	TicketDt      int64  `json:"dt,omitempty" bson:"ticket_dt"`
}
type Event struct {
	Id            int64  `json:"id,omitempty" bson:"event_id"`
	Title         string `json:"title,omitempty" bson:"show_title"`
	EventDT       int64  `json:"dt,omitempty" bson:"event_dt"`
	VenueId       int64  `json:"venue_id,omitempty" bson:"venue_id"`
	VenueTitle    string `json:"venue_title,omitempty" bson:"venue_title"`
	HallId        int64  `json:"hall_id,omitempty" bson:"hall_id"`
	Hall          string `json:"hall,omitempty" bson:"hall_title"`
	LastUpdate    int64  `json:"last_update" bson:"last_update"`
	TicketsCached int    `json:"tickets_cached" bson:"-"`
}
type EventStats struct {
	Id      int64  `json:"id,omitempty"`
	EventId int64  `json:"event_id,omitempty"`
	Dt      int64  `json:"dt,omitempty"`
	Title   string `json:"title,omitempty"`
	Sell    int64  `json:"sell,omitempty"`
	Entry   int64  `json:"entry,omitempty"`
}

type EventInfo struct {
	Tickets Tickets
	Entries Entries
}
type Tickets struct {
	Tickets int64 `json:"tickets" bson:"tickets"`
}
type Entries struct {
	Entries int64 `json:"entries" bson:"entries"`
}

type TicketsLocked struct {
	Tickets []TicketLocked
}
type TicketLocked struct {
	Barcode     string
	LockExpires int64
}

func (r *TicketsLocked) isLock(barode string) bool {
	for i := 0; i < len(r.Tickets); i++ {
		//delete old
		if r.Tickets[i].LockExpires <= time.Now().Unix() {
			r.Tickets = append(r.Tickets[:i], r.Tickets[i+1:]...)
			i--
			if i < 0 {
				return false
			}
		}
		if r.Tickets[i].Barcode == barode {
			return r.Tickets[i].LockExpires >= time.Now().Unix()
		}
	}
	return false
}
func (r *TicketsLocked) addLock(barode string) {
	if !r.isLock(barode) {
		r.Tickets = append(r.Tickets, TicketLocked{barode, time.Now().Add(time.Second * TICKET_LOCK_TIME_FOR_REENTRY).Unix()})
	}
}

type TicketsMaster struct {
	Tickets []TicketMaster
}
type TicketMaster struct {
	Barcode string
}

type Terminals struct {
	Terminals []Terminal `json:"terminals"`
}
type Groups struct {
	Groups []Group `json:"groups"`
}

func (r *Groups) BildingsIds() []int64 {
	ids := []int64{}
	for _, v := range r.Groups {
		ids = append(ids, v.BuildingId)
	}
	return ids
}
func (r *Groups) ExcludeIds() []int64 {
	ids := []int64{}
	for _, v := range r.Groups {
		for _, h := range v.Exclude_halls {
			ids = append(ids, h)
		}
	}
	return ids
}

type Events struct {
	Events []Event `json:"events"`
}

func (r *Events) EventsIds() []int64 {
	ids := []int64{}
	for _, v := range r.Events {
		ids = append(ids, v.Id)
	}
	return ids
}
func (r *Events) EventById(eventid int64) Event {
	for i := range r.Events {
		if r.Events[i].Id == eventid {
			return r.Events[i]
		}
	}
	return Event{}
}

type Entry struct {
	EventId       int64  `json:"event_id" bson:"event_id"`
	TicketBarcode string `json:"ticket_barcode" bson:"ticket_barcode"`
	TerminalId    int64  `json:"terminal_id" bson:"terminal_id"`
	OperationDt   int64  `json:"operation_dt" bson:"operation_dt"`
	ResultCode    int64  `json:"result_code" bson:"result_code"`
	Direction     string `json:"direction" bson:"direction"`
}

func (r *Entry) toAction() Action {
	return Action{r.OperationDt, r.TerminalId, r.Direction}
}

type Log struct {
	Dt      int64  `json:"dt" bson:"dt"`
	Data    string `json:"data" bson:"data"`
	Message string `json:"message" bson:"message"`
	Code    int64  `json:"code" bson:"code"`
}

type Group struct {
	Id              int64   `bson:"id" json:"id" schema:"id"`
	Name            string  `bson:"name" json:"name" schema:"name,required"`
	BuildingId      int64   `bson:"building_id" json:"building_id" schema:"building_id"`
	BuildingName    string  `bson:"building_name" json:"building_name" schema:"building_name"`
	BuildingAddress string  `bson:"building_address" json:"building_address" schema:"building_address"`
	Exclude_halls   []int64 `json:"-" bson:"exclude_halls" schema:"-"`
}
type Action struct {
	Tms        int64  `json:"tms,omitempty"`
	TerminalId int64  `json:"gate,omitempty"`
	Direction  string `json:"direction,omitempty"`
}

type SKDResponse struct {
	Result     SKDResult `json:"result"`
	Ticket     Ticket    `json:"ticket,omitempty"`
	Event      Event     `json:"event,omitempty"`
	LastAction Action    `json:"last_action,omitempty"`
}
type SKDRegistrationResponse struct {
	Result     SKDRegistrationResult `json:"result"`
	Ticket     Ticket                `json:"ticket,omitempty"`
	Event      Event                 `json:"event,omitempty"`
	LastAction Action                `json:"last_action,omitempty"`
}
type SKDResult struct {
	Code int64 `json:"code"`
}
type SKDRegistrationResult struct {
	Code  int64 `json:"code"`
	Entry bool  `json:"entry"`
	Exit  bool  `json:"exit"`
}
