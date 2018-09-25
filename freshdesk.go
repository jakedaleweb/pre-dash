package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/subosito/gotenv"
)

func getIncidentTickets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=300")
	gotenv.Load()

	ticketsJson, getTickErr := getTickets()
	if getTickErr != nil {
		log.Fatal(getTickErr)
	}

	tickets := new(Tickets)
	unmarshErr := unmarshalTickets(ticketsJson, tickets)
	if unmarshErr != nil {
		log.Fatal(unmarshErr)
	}

	incidents, filterErr := filterIncidentTickets(tickets)
	if filterErr != nil {
		log.Fatal(filterErr)
	}

	length := len(incidents)
	if length < 1 {
		w.Write([]byte("Only one Incident in the last <TIME> cannot work out average time between"))
		return
	}

	avgTimeBetween := getAverageBetween(length, incidents)
	w.Write([]byte(fmt.Sprintf("Average time between incident tickets: %v", avgTimeBetween)))
}

// Returns tickets from Freshdesk as bytes
func getTickets() ([]byte, error) {

	url := strings.Join([]string{"https://", os.Getenv("FRESHDESK_URL"), "/api/v2/tickets"}, "")
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

// Returns a list of incident tickets sorted by creation time
func filterIncidentTickets(tickets *Tickets) ([]Ticket, error) {
	var incidents []Ticket
	for _, ticket := range *tickets {
		if ticket.Type == "Incident" {
			incidents = append(incidents, ticket)
		}
	}

	sort.Slice(incidents, func(i, j int) bool {
		return incidents[i].CreatedAt.Before(incidents[j].CreatedAt)
	})

	return incidents, nil
}

// Returns average time between incident tickets creation in minutes
func getAverageBetween(length int, incidents []Ticket) float64 {
	var timeBetween float64

	for i := 1; i < length; i++ {
		prev := i - 1
		between := incidents[i].CreatedAt.Sub(incidents[prev].CreatedAt).Minutes()
		timeBetween += between
	}

	avgTimeBetween := timeBetween / float64(length)

	return avgTimeBetween
}