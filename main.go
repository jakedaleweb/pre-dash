package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/pingdom", getUptimes)
	http.HandleFunc("/freshdesk", getIncidentTickets)
	if err := http.ListenAndServe(":8082", nil); err != nil {
		panic(err)
	}
}
