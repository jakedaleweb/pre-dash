package main

import (
	"fmt"
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

	_, incidentsPrev := sortIncidentTickets(tickets)

	avg := getAverageBetween(incidentsPrev)
	if avg != 1.3429166666666665 {
		t.Error(fmt.Sprintf("Average time between incidents should be 1.3429166666666665 - %v received", avg))
	}
}
