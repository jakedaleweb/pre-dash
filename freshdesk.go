package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// Returns tickets from Freshdesk as bytes
func getTickets(viewID string) ([]byte, error) {

	// https://developers.freshdesk.com/api/#view_a_ticket
	// custom endpoint as getting all tickets seems to not work
	url := strings.Join([]string{"https://", os.Getenv("FRESHDESK_URL"), "/helpdesk/tickets/view/", viewID, "?format=json"}, "")
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(os.Getenv("FRESHDESK_TOKEN"), "x")
	resp, err := client.Do(req)

	if err != nil {
		return []byte{}, err
	}

	bodyText, err := ioutil.ReadAll(resp.Body)

	return bodyText, nil
}

// Converts json response from Freshdesk to struct
func unmarshalTickets(ticketsJson []byte, tickets *Tickets) error {
	err := json.Unmarshal(ticketsJson, tickets)
	if err != nil {
		return err
	}

	return nil
}

// Returns two lists of incident tickets sorted by creation time.
// First lot of tickets are from the previous 2 weeks the other is from the 2 preceeding that
func sortTickets(tickets *Tickets) ([]Ticket, []Ticket) {
	var incidentsRecent []Ticket
	var incidentsPrev []Ticket

	for _, ticket := range *tickets {
		if ticket.CreatedAt.After(time.Now().Add(-14 * 24 * time.Hour)) {
			incidentsRecent = append(incidentsRecent, ticket)
		} else {
			incidentsPrev = append(incidentsPrev, ticket)
		}
	}

	sort.Slice(incidentsRecent, func(i, j int) bool {
		return incidentsRecent[i].CreatedAt.Before(incidentsRecent[j].CreatedAt)
	})

	sort.Slice(incidentsPrev, func(i, j int) bool {
		return incidentsPrev[i].CreatedAt.Before(incidentsPrev[j].CreatedAt)
	})

	return incidentsRecent, incidentsPrev
}

// Returns average time between incident tickets creation in hours
func getAverageBetween(incidents []Ticket) float64 {
	var timeBetween float64

	for i := 1; i < len(incidents); i++ {
		prev := i - 1
		between := incidents[i].CreatedAt.Sub(incidents[prev].CreatedAt).Hours()
		timeBetween += between
	}

	avgTimeBetween := timeBetween / float64(len(incidents))
	return avgTimeBetween
}
