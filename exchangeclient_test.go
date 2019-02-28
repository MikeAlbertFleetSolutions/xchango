package xchango

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCalendarDetailRequest(t *testing.T) {
	config := ExchangeConfig{
		MaxFetchSize: 101,
	}

	xchang, err := NewExchangeClient(config)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	appoints := []Appointment{
		Appointment{ItemID: "alpha", ChangeKey: "123"},
		Appointment{ItemID: "beta", ChangeKey: "456"},
	}
	requestbytes, err := xchang.buildCalendarDetailRequest(appoints)
	if err != nil {
		return
	}
	request := string(requestbytes)
	assert.NotNil(t, request)

	assert.True(t, strings.Contains(request, `<typ:ItemId Id="alpha" ChangeKey="123" />`))
	assert.True(t, strings.Contains(request, `<typ:ItemId Id="beta" ChangeKey="456" />`))
}

func TestBuildCalendarItemRequest(t *testing.T) {
	config := ExchangeConfig{
		MaxFetchSize: 99,
	}

	xchang, err := NewExchangeClient(config)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	requestbytes, err := xchang.buildCalendarItemRequest("black", "ninja")
	if err != nil {
		return
	}
	request := string(requestbytes)
	assert.NotNil(t, request)

	// Only testing the two lines that get edited
	assert.True(t, strings.Contains(request, `<typ:FolderId Id="black" ChangeKey="ninja" />`))

	// Date string should always be the same length so this should always be the same
	start := strings.Index(request, "<mes:CalendarView")
	end := strings.Index(request, "<mes:ParentFolderIds")
	// Looks somethign like <mes:CalendarView MaxEntriesReturned="100" StartDate="2015-04-21T05:59:57Z" EndDate="2015-05-05T05:59:57Z"/>
	calendarline := request[start:end]
	keyvaluepairs := strings.Split(calendarline, " ")

	// Verify the dates are there and the max entries contains a number
	count := 0
	for _, keyvalue := range keyvaluepairs {
		if strings.Index(keyvalue, "=") > 0 {
			parts := strings.Split(keyvalue, "=")
			assert.NotNil(t, parts)
			switch parts[0] {
			case "MaxEntriesReturned":
				numstring := parts[1][1 : len(parts[1])-1]
				i, err := strconv.ParseInt(numstring, 0, 64)
				assert.Nil(t, err)
				assert.Equal(t, int64(99), i)
				count |= 1
			case "StartDate":
				assert.Equal(t, 22, len(parts[1]))
				count |= 2
			case "EndDate":
				assert.Equal(t, 25, len(parts[1])) // Length includes ending xml />
				count |= 4
			}
		}
	}
	assert.Equal(t, 7, count, "All properties were not found in the reponse.  Total: "+string(count))
}
