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

	incidents, filterErr := sortIncidentTickets(tickets)
	if filterErr != nil {
		log.Fatal(filterErr)
	}

	avg := getAverageBetween(len(incidents), incidents)
	if avg != "1.343" {
		t.Error(fmt.Sprintf("Average time between incidents should be 1.343 - %v received", avg))
	}
}
