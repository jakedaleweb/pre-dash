package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/pingdom", getUptimes)
	if err := http.ListenAndServe(":8082", nil); err != nil {
		panic(err)
	}
}
