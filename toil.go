package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

func showMacTickets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=300")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	ticketsJson, getTickErr := getTickets(os.Getenv("MAC_VIEW_ID"))
	if getTickErr != nil {
		log.Fatal(getTickErr)
	}

	tickets := new(Tickets)
	unmarshErr := unmarshalTickets(ticketsJson, tickets)
	if unmarshErr != nil {
		log.Fatal(unmarshErr)
	}

	toilRecent, toilPrev := sortTickets(tickets)

	// for example if recent is count 5 and prev is 6 then diff is -1 which is good
	// i.e. one less ticket
	diff := len(toilRecent) - len(toilPrev)
	increase := true
	if diff < 0 {
		increase = false
	}

	tmpl := template.Must(template.ParseFiles("templates/toil.html"))
	data := ToilPage{
		Title:    "Toil",
		Count:    len(toilRecent),
		Diff:     diff,
		Increase: increase,
	}
	tmpl.Execute(w, data)
}
