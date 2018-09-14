package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/subosito/gotenv"
)

func getIncidentTickets(w http.ResponseWriter, r *http.Request) {
	gotenv.Load()

	url := strings.Join([]string{"https://", os.Getenv("FRESHDESK_URL"), "/api/v2/tickets"}, "")
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(os.Getenv("FRESHDESK_TOKEN"), "x")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)

	s := string(bodyText)
	for _, ticket := range s {
		fmt.Println(string(ticket))
		break
	}
	w.Write([]byte(s))
}
