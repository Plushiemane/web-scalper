package main

import (
	"net/http"
)

func main() {

	//praktykant-stażysta
	// + 1,2,3,4,5, etc.
	//fmt.Println("Total pages:", findNumberPages(doc))

	http.HandleFunc("/jobs", jobsHandler)
	http.ListenAndServe(":8080", nil)
}
