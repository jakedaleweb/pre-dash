package main

import (
	"html/template"
	"net/http"

	"github.com/subosito/gotenv"
)

func main() {
	gotenv.Load()

	http.HandleFunc("/pingdom", getUptimes)
	http.HandleFunc("/freshdesk", getIncidentTickets)
	http.HandleFunc("/", homePage)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	if err := http.ListenAndServe(":8082", nil); err != nil {
		panic(err)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=300")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	data := HomePage{Title: "Home"}
	tmpl.Execute(w, data)
}
