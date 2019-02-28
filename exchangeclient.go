package xchango

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"text/template"
	"time"

	httpntlm "github.com/vadimi/go-http-ntlm"
)

// ExchangeUser defines the NTLM authentication parameters to connect to the Exchange server
type ExchangeUser struct {
	Username string
	Password string
	Domain   string
}

// ExchangeConfig defines all the parameters for how we interact with the Exchange server
type ExchangeConfig struct {
	ExchangeUser
	ExchangeURL   string
	MaxFetchSize  int
	LookAheadDays int
}

type exchangeVersion interface {
	FolderRequest() string
	CalendarRequest() string
	CalendarDetailRequest() string
}

// ExchangeCalendar definition
type ExchangeCalendar struct {
	Folderid  string
	Changekey string
}

// ExchangeClient the callers reference to us
type ExchangeClient struct {
	exchangeConfig ExchangeConfig
	version        exchangeVersion
	client         http.Client
}

// NewExchangeClient creates a new ExchangeClient
func NewExchangeClient(config ExchangeConfig) (ec *ExchangeClient, err error) {
	// setup ews http client
	var cookieJar *cookiejar.Jar
	cookieJar, err = cookiejar.New(nil)
	if err != nil {
		return
	}

	httpclient := http.Client{
		Jar:     cookieJar,
		Timeout: time.Second * 30,
		Transport: &httpntlm.NtlmTransport{
			Domain:   config.ExchangeUser.Domain,
			User:     config.ExchangeUser.Username,
			Password: config.ExchangeUser.Password,
		},
	}

	ec = &ExchangeClient{
		exchangeConfig: config,
		version:        exchange2006{},
		client:         httpclient,
	}

	return
}

// GetCalendar gets the users exchange calendar folder
func (ec *ExchangeClient) GetCalendar() (cal *ExchangeCalendar, err error) {
	// make request
	soapReq := ec.version.FolderRequest()
	var results string
	results, err = ec.postContents([]byte(soapReq))
	if err != nil {
		return
	}

	// parse returned soap
	item := parseCalendarFolder(string(results))

	cal = &ExchangeCalendar{
		Folderid:  item.ID,
		Changekey: item.ChangeKey,
	}

	return
}

// GetAppointments gets users appointments
func (ec *ExchangeClient) GetAppointments(cal *ExchangeCalendar) (appointments []Appointment, err error) {
	// just get ids for each appointment
	calRequest, err := ec.buildCalendarItemRequest(cal.Folderid, cal.Changekey)
	if err != nil {
		return
	}
	calResults, err := ec.postContents(calRequest)
	if err != nil {
		return
	}
	itemIds := parseAppointments(calResults)

	// get all the fields given the ids
	appRequest, err := ec.buildCalendarDetailRequest(itemIds)
	if err != nil {
		return
	}
	appResults, err := ec.postContents(appRequest)
	if err != nil {
		return
	}

	// unmarshall
	appointments = parseAppointments(appResults)
	if err != nil {
		return nil, err
	}

	return
}

func (ec *ExchangeClient) buildCalendarItemRequest(folderid string, changekey string) (request []byte, err error) {
	days := ec.exchangeConfig.LookAheadDays
	if days < 1 {
		days = 7
	}
	startDate := time.Now().UTC().Format(time.RFC3339)
	endDate := time.Now().UTC().AddDate(0, 0, days).Format(time.RFC3339)

	data := struct {
		StartDate    string
		EndDate      string
		FolderID     string
		ChangeKey    string
		MaxFetchSize int
	}{
		startDate,
		endDate,
		folderid,
		changekey,
		ec.exchangeConfig.MaxFetchSize,
	}

	t, err := template.New("cal").Parse(ec.version.CalendarRequest())
	if err != nil {
		return
	}

	var doc bytes.Buffer
	t.Execute(&doc, data)
	if err != nil {
		return
	}

	request = doc.Bytes()

	return
}

func (ec *ExchangeClient) buildCalendarDetailRequest(itemIds []Appointment) (request []byte, err error) {
	data := struct {
		Appointments []Appointment
	}{
		itemIds,
	}

	t, err := template.New("detail").Parse(ec.version.CalendarDetailRequest())
	if err != nil {
		return
	}

	var doc bytes.Buffer
	t.Execute(&doc, data)
	request = doc.Bytes()

	return
}

func (ec *ExchangeClient) postContents(contents []byte) (data string, err error) {
	request, err := http.NewRequest("POST", ec.exchangeConfig.ExchangeURL, bytes.NewBuffer(contents))
	if err != nil {
		return
	}
	request.Header.Set("Accept", "text/xml")
	request.Header.Set("Content-Type", "text/xml")

	response, err := ec.client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()

	var content []byte
	content, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	data = string(content)

	return
}
