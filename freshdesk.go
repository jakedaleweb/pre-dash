package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

func getIncidentTickets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=300")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	ticketsJson, getTickErr := getTickets()
	if getTickErr != nil {
		log.Fatal(getTickErr)
	}

	tickets := new(Tickets)
	unmarshErr := unmarshalTickets(ticketsJson, tickets)
	if unmarshErr != nil {
		log.Fatal(unmarshErr)
	}

	incidents, filterErr := sortIncidentTickets(tickets)
	if filterErr != nil {
		log.Fatal(filterErr)
	}

	length := len(incidents)
	if length < 1 {
		w.Write([]byte("Only one Incident in the last 30 days cannot work out average time between"))
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/freshdesk.html"))
	data := FreshdeskPage{
		Title: "Time between incidents",
		Avg:   getAverageBetween(length, incidents),
		Count: length,
	}
	tmpl.Execute(w, data)
}

// Returns tickets from Freshdesk as bytes
func getTickets() ([]byte, error) {

	// https://developers.freshdesk.com/api/#view_a_ticket
	// custom endpoint as getting all tickets seems to not work
	url := strings.Join([]string{"https://", os.Getenv("FRESHDESK_URL"), "/helpdesk/tickets/view/206953?format=json"}, "")
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
func sortIncidentTickets(tickets *Tickets) ([]Ticket, error) {
	var incidents []Ticket
	for _, ticket := range *tickets {
		incidents = append(incidents, ticket)
	}

	sort.Slice(incidents, func(i, j int) bool {
		return incidents[i].CreatedAt.Before(incidents[j].CreatedAt)
	})

	return incidents, nil
}

// Returns average time between incident tickets creation in hours
func getAverageBetween(length int, incidents []Ticket) string {
	var timeBetween float64

	for i := 1; i < length; i++ {
		prev := i - 1
		between := incidents[i].CreatedAt.Sub(incidents[prev].CreatedAt).Hours()
		timeBetween += between
	}

	avgTimeBetween := timeBetween / float64(length)
	strAvg := strconv.FormatFloat(avgTimeBetween, 'f', 3, 64)

	return strAvg
}
