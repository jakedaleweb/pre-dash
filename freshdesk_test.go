package main

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestUnmarshalTickets(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/tickets.json")
	if err != nil {
		log.Fatal(err)
	}

	tickets := new(Tickets)
	unmarshErr := unmarshalTickets(b, tickets)

	if unmarshErr != nil {
		log.Fatal(unmarshErr)
	}

	incidents, filterErr := filterIncidentTickets(tickets)
	if filterErr != nil {
		log.Fatal(filterErr)
	}

	for _, incident := range incidents {
		if incident.Type != "Incident" {
			t.Error("Ticket type should be incident")
		}
	}

	avg := getAverageBetween(len(incidents), incidents)
	if avg != 74.13333333333334 {
		t.Error("Average time between incidents should be 74.13333333333334")
	}
}
