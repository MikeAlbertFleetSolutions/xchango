package xchango

import (
	"bytes"
	"encoding/xml"
	"log"
	"time"
)

type Organizer struct {
	Mailbox Mailbox
}

type Mailbox struct {
	Name string
}

type ItemID struct {
	ID        string `xml:"Id,attr"`
	ChangeKey string `xml:"ChangeKey,attr"`
}

type CalendarItem struct {
	ItemID         ItemID `xml:"ItemId"`
	Subject        string
	DisplayCc      string
	DisplayTo      string
	Start          string
	End            string
	IsAllDayEvent  bool
	Location       string
	MyResponseType string
	Organizer      Organizer
	Body           Body
}

type Body struct {
	BodyType string `xml:"BodyType,attr"`
	Body     string `xml:",chardata"`
}

type Appointment struct {
	ItemID         string
	ChangeKey      string
	Subject        string
	Cc             string
	To             string
	Start          time.Time
	End            time.Time
	IsAllDayEvent  bool
	Location       string
	MyResponseType string
	Organizer      string
	Body           string
	BodyType       string
}

func parseCalendarFolder(soap string) ItemID {
	decoder := xml.NewDecoder(bytes.NewBufferString(soap))
	var itemID ItemID

	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "FolderId" {
				decoder.DecodeElement(&itemID, &se)
				break
			}
		}
	}
	return itemID
}

func parseAppointments(soap string) []Appointment {
	decoder := xml.NewDecoder(bytes.NewBufferString(soap))

	appointments := make([]Appointment, 0)

	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "CalendarItem" {
				var item CalendarItem
				decoder.DecodeElement(&item, &se)
				appointments = append(appointments, item.toAppointment())
			}
		}
	}
	return appointments
}

func (c CalendarItem) toAppointment() Appointment {
	app := Appointment{
		ItemID:         c.ItemID.ID,
		ChangeKey:      c.ItemID.ChangeKey,
		Subject:        c.Subject,
		Cc:             c.DisplayCc,
		To:             c.DisplayTo,
		IsAllDayEvent:  c.IsAllDayEvent,
		Location:       c.Location,
		MyResponseType: c.MyResponseType,
		Organizer:      c.Organizer.Mailbox.Name,
		Body:           c.Body.Body,
		BodyType:       c.Body.BodyType,
	}
	if len(c.Start) > 0 {
		t1, err := time.Parse(time.RFC3339, c.Start)
		if err != nil {
			log.Printf("Error while parsing time.  Start time string was: %v", c.Start)
			log.Println(err)
		}
		app.Start = t1
	}

	if len(c.End) > 0 {
		t1, err := time.Parse(time.RFC3339, c.End)
		if err != nil {
			log.Printf("Error while parsing time.  End time string was: %v", c.End)
			log.Println(err)
		}
		app.End = t1
	}
	return app
}
