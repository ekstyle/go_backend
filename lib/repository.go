package lib

import (
	"crypto/md5"
	"encoding/hex"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"os"
	"time"
)

type Repository struct {
	Server   string
	Database string
}

const SALT = "1c2cf9a0a9031262b894fac41f05e656"
const USER_COLLECTION = "users"
const TERMINALS_COLLECTION = "terminals"
const GROUPS_COLLECTION = "groups"
const TICKETS_COLLECTION = "tickets"
const EVENTS_COLLECTION = "events"
const ENTRY_COLLECTION = "entry"

var db *mgo.Database

func (r *Repository) Connect() {
	r.Server = os.Getenv("MONGO_URL")
	r.Database = os.Getenv("MONGO_DB")
	session, err := mgo.Dial(r.Server)
	if err != nil {
		log.Fatal(err)
	}
	session.SetMode(mgo.Monotonic, true)
	log.Println("Connected to ", r.Server, "with", r.Database, "database.")
	db = session.DB(r.Database)
	// Optional. Switch the session to a monotonic behavior.

}
func getResultForEntry(entryItem Entry) (entry bool, exit bool) {
	if entryItem == (Entry{}) || entryItem.Direction == "exit" {
		return true, false
	}
	return false, true
}
func hashPassword(pass string) string {
	hash := md5.New()
	hash.Write([]byte(pass + SALT))
	return hex.EncodeToString(hash.Sum(nil))
}
func (r *Repository) CheckUser(userLogin UserLogin) (bool, *Exception) {
	//result := &User{}
	//db.C("users").Insert(&User{"tester","just test"})
	userCount, errFind := db.C("users").Find(bson.M{"active": true, "login": userLogin.Login, "password": hashPassword(userLogin.Password)}).Count()
	if errFind != nil {
		return false, &Exception{CANT_SELECT_EXEPTION, errFind.Error()}
	}
	//Correct user
	if userCount == 1 {
		return true, nil
	}
	//Not found
	return false, nil
}
func (r *Repository) Terminals() interface{} {
	/*	query := []bson.M{{
			"$lookup": bson.M{
				"from": "groups",
				"localField": "groups",
				"foreignField": "id",
				"as": "groups",
			}}}
		terms :=[]Terminal{}
		pipe := db.C(TERMINALS_COLLECTION).Pipe(query)
		pipe.All(&terms)*/
	var terms []Terminal
	db.C(TERMINALS_COLLECTION).Find(nil).All(&terms)

	return Terminals{terms}
}
func (r *Repository) Groups() interface{} {
	var result []Group
	db.C(GROUPS_COLLECTION).Find(nil).All(&result)

	return Groups{result}
}
func (r *Repository) AddUser(user User) *Exception {
	//Try to find user
	userCount, errFind := db.C(USER_COLLECTION).Find(bson.M{"login": user.Login}).Count()
	if errFind != nil {
		return &Exception{CANT_SELECT_EXEPTION, errFind.Error()}
	}
	if userCount > 0 {
		return &Exception{USER_EXIST_EXEPTION, ""}
	}
	user.Password = hashPassword(user.Password)
	user.Active = true
	errInsert := db.C(USER_COLLECTION).Insert(user)
	if errInsert != nil {
		return &Exception{CANT_INSERT_EXEPTION, errInsert.Error()}
	}
	return nil
}
func (r *Repository) SetGroup(group Group) *Exception {

	db.C(GROUPS_COLLECTION).Upsert(bson.M{"name": group.Name}, group)
	return nil
}

func (r *Repository) SetTerminal(terminal Terminal) *Exception {
	log.Println(terminal)
	db.C(TERMINALS_COLLECTION).Update(bson.M{"id": terminal.Id}, bson.M{"$set": terminal})
	return nil
}
func (r *Repository) AddGroup(group Group) *Exception {
	//Try to find user
	groupCount, errFind := db.C(GROUPS_COLLECTION).Find(bson.M{"name": group.Name}).Count()
	if errFind != nil {
		return &Exception{CANT_SELECT_EXEPTION, errFind.Error()}
	}
	if groupCount > 0 {
		return &Exception{USER_EXIST_EXEPTION, ""}
	}

	errInsert := db.C(GROUPS_COLLECTION).Insert(group)
	if errInsert != nil {
		return &Exception{CANT_INSERT_EXEPTION, errInsert.Error()}
	}
	return nil
}
func (r *Repository) RemoveGroup(group Group) *Exception {

	db.C(GROUPS_COLLECTION).Remove(group)

	return nil
}
func (r *Repository) SyncEvent(eventId int) *Exception {
	eventExport := api.GetEventACS(eventId)
	//sync Event
	db.C(EVENTS_COLLECTION).Upsert(bson.M{"event_id": eventExport.Content.Data.Event.EventID}, eventExport.Content.Data.Event)
	//sync Tickets
	bulk := db.C(TICKETS_COLLECTION).Bulk()
	timeUnix := time.Now().Unix()
	source := api.Source()
	for _, element := range eventExport.Content.Data.Event.Tickets {
		element.EventID = eventExport.Content.Data.Event.EventID
		element.LastUpdate = timeUnix
		element.Source = source
		bulk.Upsert(bson.M{"ticket_id": element.TicketID, "event_id": element.EventID}, element)
	}
	_, err := bulk.Run()
	//remove old items
	if err == nil {
		db.C(TICKETS_COLLECTION).RemoveAll(bson.M{"event_id": eventExport.Content.Data.Event.EventID, "source": source, "last_update": bson.M{"$lt": timeUnix}})
	}
	return nil
}
func (r *Repository) ValidateTicket(barcode string, term Terminal) (SKDResponse, *Exception) {
	curentGroups := r.GetGroupsByTerminal(term)
	currentEvents := r.GetActiveEventsByGroups(curentGroups)
	log.Println(currentEvents.EventsIds())
	ticket := Ticket{}
	db.C(TICKETS_COLLECTION).Find(bson.M{"ticket_barcode": barcode, "event_id": bson.M{"$in": currentEvents.EventsIds()}}).One(&ticket)

	if (Ticket{}) != ticket {
		entry := r.CheckTicketForEntry(ticket)

		return SKDResponse{SKDResult{ENTRY_RESULT_CODE_ACCEPT}, ticket, currentEvents.EventById(ticket.EventId), entry.toAction()}, nil
	}
	//Not Found
	ticket.TicketBarcode = barcode
	return SKDResponse{SKDResult{ENTRY_RESULT_CODE_NOTFOUND}, ticket, Event{}, Action{}}, nil

}
func (r *Repository) ValidateRegistrateTicket(barcode string, term Terminal, direction string) (SKDRegistrationResponse, *Exception) {
	curentGroups := r.GetGroupsByTerminal(term)
	currentEvents := r.GetActiveEventsByGroups(curentGroups)
	log.Println(currentEvents.EventsIds())
	ticket := Ticket{}
	db.C(TICKETS_COLLECTION).Find(bson.M{"ticket_barcode": barcode, "event_id": bson.M{"$in": currentEvents.EventsIds()}}).One(&ticket)

	if (Ticket{}) != ticket {
		entryItem := r.CheckTicketForEntry(ticket)
		entry, exit := getResultForEntry(entryItem)
		if entry && direction == "entry" {
			//Entry allowed
			return SKDRegistrationResponse{SKDRegistrationResult{ENTRY_RESULT_CODE_ACCEPT, entry, exit}, ticket, currentEvents.EventById(ticket.EventId), entryItem.toAction()}, nil
		}
		if exit && direction == "exit" {
			//Exit
			return SKDRegistrationResponse{SKDRegistrationResult{ENTRY_RESULT_CODE_ACCEPT, entry, exit}, ticket, currentEvents.EventById(ticket.EventId), entryItem.toAction()}, nil
		}
		//reentry
		entryRecord := Entry{ticket.EventId, ticket.TicketBarcode, term.Id, time.Now().Unix(), ENTRY_RESULT_CODE_REENTRY, direction}
		errInsert := db.C(ENTRY_COLLECTION).Insert(entryRecord)
		if errInsert != nil {
			return SKDRegistrationResponse{}, &Exception{CANT_INSERT_EXEPTION, errInsert.Error()}
		}
		return SKDRegistrationResponse{SKDRegistrationResult{ENTRY_RESULT_CODE_REENTRY, false, false}, ticket, currentEvents.EventById(ticket.EventId), entryItem.toAction()}, nil
	}
	//Not Found
	ticket.TicketBarcode = barcode
	return SKDRegistrationResponse{SKDRegistrationResult{ENTRY_RESULT_CODE_NOTFOUND, false, false}, ticket, Event{}, Action{}}, nil
}
func (r *Repository) RegistrateTicket(barcode string, term Terminal, direction string) (SKDResult, *Exception) {
	curentGroups := r.GetGroupsByTerminal(term)
	currentEvents := r.GetActiveEventsByGroups(curentGroups)
	log.Println(currentEvents.EventsIds())
	ticket := Ticket{}
	db.C(TICKETS_COLLECTION).Find(bson.M{"ticket_barcode": barcode, "event_id": bson.M{"$in": currentEvents.EventsIds()}}).One(&ticket)

	if (Ticket{}) != ticket {
		entryItem := r.CheckTicketForEntry(ticket)
		entry, exit := getResultForEntry(entryItem)
		if entry && direction == "entry" {
			//Entry allowed
			entryRecord := Entry{ticket.EventId, ticket.TicketBarcode, term.Id, time.Now().Unix(), ENTRY_RESULT_CODE_ACCEPT, direction}
			errInsert := db.C(ENTRY_COLLECTION).Insert(entryRecord)
			if errInsert != nil {
				return SKDResult{ENTRY_RESULT_CODE_NOTFOUND}, &Exception{CANT_INSERT_EXEPTION, errInsert.Error()}
			}
			return SKDResult{ENTRY_RESULT_CODE_ACCEPT}, nil
		}
		if exit && direction == "exit" {
			//Exit
			entryRecord := Entry{ticket.EventId, ticket.TicketBarcode, term.Id, time.Now().Unix(), ENTRY_RESULT_CODE_ACCEPT, direction}
			errInsert := db.C(ENTRY_COLLECTION).Insert(entryRecord)
			if errInsert != nil {
				return SKDResult{ENTRY_RESULT_CODE_NOTFOUND}, &Exception{CANT_INSERT_EXEPTION, errInsert.Error()}
			}
			return SKDResult{ENTRY_RESULT_CODE_ACCEPT}, nil
		}
		//reentry
		entryRecord := Entry{ticket.EventId, ticket.TicketBarcode, term.Id, time.Now().Unix(), ENTRY_RESULT_CODE_REENTRY, direction}
		errInsert := db.C(ENTRY_COLLECTION).Insert(entryRecord)
		if errInsert != nil {
			return SKDResult{ENTRY_RESULT_CODE_NOTFOUND}, &Exception{CANT_INSERT_EXEPTION, errInsert.Error()}
		}
		return SKDResult{ENTRY_RESULT_CODE_NOTFOUND}, nil
	}
	//Not Found
	return SKDResult{ENTRY_RESULT_CODE_NOTFOUND}, nil
}
func (r *Repository) GetGroupsByTerminal(terminal Terminal) Groups {
	groups := Groups{}
	db.C(GROUPS_COLLECTION).Find(bson.M{"id": bson.M{"$in": terminal.Groups}}).All(&groups.Groups)
	return groups
}
func (r *Repository) GetActiveEventsByGroups(groups Groups) Events {
	events := Events{}
	db.C(EVENTS_COLLECTION).Find(bson.M{"venue_id": bson.M{"$in": groups.BildingsIds()}}).All(&events.Events)
	return events
}
func (r *Repository) GetTerminalById(terminalId int64) Terminal {
	term := Terminal{}
	db.C(TERMINALS_COLLECTION).Find(bson.M{"id": terminalId}).One(&term)
	return term
}
func (r *Repository) GetAuthTerminalById(terminalId int64) AuthStruct {
	term := AuthStruct{}
	db.C(TERMINALS_COLLECTION).Find(bson.M{"id": terminalId}).One(&term.Auth)
	return term
}
func (r *Repository) CheckTicketForEntry(ticket Ticket) Entry {
	entry := Entry{}
	db.C(ENTRY_COLLECTION).Find(bson.M{"ticket_barcode": ticket.TicketBarcode, "event_id": ticket.EventId, "result_code": ENTRY_RESULT_CODE_ACCEPT}).Sort("-operation_dt").One(&entry)
	return entry
}
