package speeddaemon

import (
	"log"
	"math"
	"math/rand"
	"slices"
	"sort"
)

type Message interface {
	Process(cd *CentralDispatcher)
}

type CentralDispatcher struct {
	MessageQueue chan Message
	Records      map[U16]map[Str][]*Record // Roads -> License -> List of Records
	Dispatchers  map[U16][]chan<- ClientMessage
	SpeedLimits  map[U16]U16
	Tickets      map[Str][]U32
}

type Record struct {
	Mile U16
	Time U32
}

type RegisterCamera struct {
	Road  U16
	Limit U16
}

type RegisterDispatcher struct {
	Roads   []U16
	Channel chan<- ClientMessage
}

type Observation struct {
	Road      U16
	License   Str
	Mile      U16
	Timestamp U32
}

func NewCentralDispatcher() *CentralDispatcher {
	return &CentralDispatcher{
		MessageQueue: make(chan Message, 100),
		Records:      make(map[U16]map[Str][]*Record),
		Dispatchers:  make(map[U16][]chan<- ClientMessage),
		SpeedLimits:  make(map[U16]U16),
		Tickets:      make(map[Str][]U32),
	}
}

func (cd *CentralDispatcher) Start() {
	for msg := range cd.MessageQueue {
		msg.Process(cd)
	}
}

func (rd *RegisterDispatcher) Process(cd *CentralDispatcher) {
	log.Println("Registering dispatcher with roads", rd.Roads)
	for _, road := range rd.Roads {
		if _, exists := cd.Dispatchers[road]; !exists {
			cd.Dispatchers[road] = make([]chan<- ClientMessage, 0)
		}
		// Add the dispatcher channel to the list of dispatchers for the road
		cd.Dispatchers[road] = append(cd.Dispatchers[road], rd.Channel)
	}
	// We may have to send a ticket, so check all existing records
	for _, road := range rd.Roads {
		for license, records := range cd.Records[road] {
			cd.calculateTickets(road, license, records)
		}
	}
}

func (rc *RegisterCamera) Process(cd *CentralDispatcher) {
	log.Println("Registering camera on road: ", rc.Road, "with limit: ", rc.Limit)
	cd.SpeedLimits[rc.Road] = rc.Limit
	if _, exists := cd.Records[rc.Road]; !exists {
		cd.Records[rc.Road] = make(map[Str][]*Record, 0)
	}
}

func (o *Observation) Process(cd *CentralDispatcher) {
	log.Println("Processing observation. Road: ", o.Road, "License: ", o.License, "Mile: ", o.Mile, "Timestamp: ", o.Timestamp)
	roadRecords, exists := cd.Records[o.Road]
	if !exists {
		log.Printf("Could not register observation. Road %d not registered", o.Road)
		return
	}

	licenseRecords, exists := roadRecords[o.License]
	if !exists {
		licenseRecords = []*Record{}
	}

	newRecord := &Record{
		Mile: o.Mile,
		Time: o.Timestamp,
	}

	// Append and sort
	licenseRecords = append(licenseRecords, newRecord)
	sort.Slice(licenseRecords, func(i, j int) bool {
		return licenseRecords[i].Time < licenseRecords[j].Time
	})

	// Save back to the map
	roadRecords[o.License] = licenseRecords
	cd.Records[o.Road] = roadRecords

	// Process tickets
	cd.calculateTickets(o.Road, o.License, licenseRecords)
}

func (cd *CentralDispatcher) calculateTickets(road U16, license Str, licenseRecords []*Record) {
	log.Printf("Calculating tickets for license %s on road %d", license, road)
	speedLimit := cd.SpeedLimits[road] * 100
	for i := 1; i < len(licenseRecords); i++ {
		lastRecord := *licenseRecords[i-1]
		currentRecord := *licenseRecords[i]
		speed := CalculateSpeed(lastRecord, currentRecord)
		log.Println("Calculating speed. Speed limit", speedLimit, "Speed: ", speed)
		if speed > speedLimit {
			cd.generateTicket(license, road, lastRecord, currentRecord, speed)
		}
	}
}

func CalculateSpeed(r1 Record, r2 Record) U16 {
	distance := float64(r2.Mile) - float64(r1.Mile)
	timeHours := (float64(r2.Time) - float64(r1.Time)) / 3600.0
	speedMph := distance / timeHours
	speedHundredths := math.Abs(speedMph * 100)
	return U16(speedHundredths)
}

func (cd *CentralDispatcher) generateTicket(license Str, road U16, r1 Record, r2 Record, speed U16) {
	startDay := r1.Time / 86400
	endDay := r2.Time / 86400
	// Check if the ticket is already in the list from the same day
	for i := startDay; i <= endDay; i++ {
		if slices.Contains(cd.Tickets[license], i) {
			return
		}
	}
	// Check if the dispatchers are available
	dispatchers := cd.Dispatchers[road]
	if len(dispatchers) == 0 {
		return
	}
	log.Println("Generating ticket for license: ", license, "Road: ", road, "Speed: ", speed)
	ticket := &TicketMessage{
		Plate:        license,
		Road:         road,
		MileOne:      r1.Mile,
		TimeStampOne: r1.Time,
		MileTwo:      r2.Mile,
		TimeStampTwo: r2.Time,
		Speed:        speed,
	}
	// Send the ticket to the random dispatcher
	log.Println("Dispatching ticket to random dispatcher for license: ", license, " on day: ")
	random := rand.Intn(len(dispatchers))
	dispatchers[random] <- ticket
	for i := startDay; i <= endDay; i++ {
		cd.Tickets[license] = append(cd.Tickets[license], i)
	}
}
