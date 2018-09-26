package main

import (
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
)

func showIncidentTickets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=300")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	ticketsJson, getTickErr := getTickets(os.Getenv("INCIDENT_VIEW_ID"))
	if getTickErr != nil {
		log.Fatal(getTickErr)
	}

	tickets := new(Tickets)
	unmarshErr := unmarshalTickets(ticketsJson, tickets)
	if unmarshErr != nil {
		log.Fatal(unmarshErr)
	}

	incidentsRecent, incidentsPrev := sortTickets(tickets)

	recentAvg := getAverageBetween(incidentsRecent)
	prevAvg := getAverageBetween(incidentsPrev)

	strAvg := strconv.FormatFloat(recentAvg, 'f', 3, 64)
	if math.IsNaN(recentAvg) {
		strAvg = ""
	}

	// If recent avg is 5 and previous is 6 we are down one hour so the diff is -1 which is negative
	diff := recentAvg - prevAvg
	strDiff := strconv.FormatFloat(diff, 'f', 3, 64)
	if math.IsNaN(diff) {
		strDiff = ""
	}

	// Check if the diff is less than 0 (i.e. not increased)
	increase := true
	if diff < 0 {
		increase = false
	}

	tmpl := template.Must(template.ParseFiles("templates/incidents.html"))
	data := IncidentPage{
		Title:    "Time between incidents",
		Avg:      strAvg,
		Count:    len(incidentsRecent),
		Diff:     strDiff,
		Increase: increase,
	}
	tmpl.Execute(w, data)
}
