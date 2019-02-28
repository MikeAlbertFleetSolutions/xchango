# xchango

Read calendar information from an Exchange Server using Go.  Forked from [sgoertzen/xchango](https://github.com/sgoertzen/xchango).  Changed to use NTLM for authentication and how caller interacts with Exchange Client.

# Example
```golang
package main

import (
	"log"
	"time"

	"github.com/MikeAlbertFleetSolutions/xchango"
)

func main() {
	xchangconfig := xchango.ExchangeConfig{
		ExchangeUser: xchango.ExchangeUser{
			Domain:   "oz",
			Username: "big.kahuna",
			Password: "charge!",
		},
		MaxFetchSize: 5,
		ExchangeURL:  "https://mail.oz.com/EWS/Exchange.asmx",
	}

	xchang, err := xchango.NewExchangeClient(xchangconfig)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	cal, err := xchang.GetCalendar()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	appointments, err := xchang.GetAppointments(cal)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	newYork, err := time.LoadLocation("America/New_York")
	for _, appointment := range *appointments {
		log.Printf("APPOINTMENT %+v %+v\n", appointment.Start.In(newYork), appointment.Subject)
	}
}
```