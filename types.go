package main

import (
	"time"

	"github.com/russellcardullo/go-pingdom/pingdom"
)

type PlatformType string

const (
	PlatformSSP PlatformType = "SSP"
	PlatformCWP              = "CWP"
)

type Tickets []Ticket

type Ticket struct {
	CreatedAt     time.Time `json:"created_at"`
	Deleted       bool      `json:"deleted"`
	GroupID       int64     `json:"group_id"`
	ID            int64     `json:"id"`
	Priority      int       `json:"priority"`
	RequesterID   int64     `json:"requester_id"`
	ResponderID   int64     `json:"responder_id"`
	Source        int       `json:"source"`
	Status        int       `json:"status"`
	Subject       string    `json:"subject"`
	TicketType    string    `json:"ticket_type"`
	UpdatedAt     time.Time `json:"updated_at"`
	RequesterName string    `json:"requester_name"`
	ResponderName string    `json:"responder_name"`
	ProductID     int64     `json:"product_id"`
}

type summaryOutageJsonResponse struct {
	Summary struct {
		States []State `json:"states"`
	} `json:"summary"`
}

type State struct {
	Status   string `json:"status"`
	Timefrom int64  `json:"timefrom"`
	Timeto   int64  `json:"timeto"`
}

type UptimeResult struct {
	platform PlatformType
	check    pingdom.CheckResponse
	uptime   float64
	up       int64
	down     int64
	total    int64
	unknown  int64
}

type HomePage struct {
	Title string
	Body  []byte
}

type PingdomPage struct {
	Title       string
	CwpRes      []ResultRow
	SspRes      []ResultRow
	SspUptime   string
	CwpUptime   string
	SspDiff     string
	SspIncrease bool
	CwpDiff     string
	CwpIncrease bool
}

type IncidentPage struct {
	Title    string
	Avg      string
	Count    int
	Diff     string
	Increase bool
}

type ToilPage struct {
	Title    string
	Count    int
	Diff     int
	Increase bool
}

type ResultRow struct {
	Availability string
	Name         string
	Downtime     string
	ErrorBudget  string
}
