package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Job struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

type Request struct {
	StartURL string `json:"starturl"`
}

func checkerror(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func extractJobs(doc *goquery.Document, url string) []Job {
	var jobs []Job
	pages := findNumberPages(doc)
	appendix := "&pn="
	seen := make(map[string]bool)

	for i := 1; i <= pages; i++ {
		var new_url string
		if i == 1 {
			new_url = url
		} else {
			new_url = url + appendix + strconv.Itoa(i)
		}
		//fmt.Println("Fetching:", new_url)
		pageDoc := getObj(new_url)
		count := 0
		pageDoc.Find("div.tiles_b18pwp01.core_po9665q").Each(func(_ int, s *goquery.Selection) {
			count++
			title := strings.TrimSpace(s.Find("h2.tiles_h1p4o5k6").Text())
			link, _ := s.Find("a.tiles_cnb3rfy.core_n194fgoq").Attr("href")
			if link != "" && !seen[link] {
				jobs = append(jobs, Job{Title: title, Link: link})
				seen[link] = true
			}
		})
		//fmt.Printf("Page %d: found %d jobs\n", i, count)
	}
	return jobs
}

func getObj(url string) *goquery.Document {

	response, error := http.Get(url);
	checkerror(error)

	defer response.Body.Close()
	if response.StatusCode > 400 {
		fmt.Println("Status code:", response.StatusCode)
	}

	doc, error := goquery.NewDocumentFromReader(response.Body)
	checkerror(error);

	return doc
}

func findNumberPages(doc *goquery.Document) int {
	maxPage := doc.Find(`span[data-test="top-pagination-max-page-number"]`).Text()
	number, _ := strconv.Atoi(maxPage)
	return number
}

func jobsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
    
    // Handle preflight OPTIONS request
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	doc := getObj(req.StartURL)
	if doc == nil {
		http.Error(w, "Failed to fetch starturl", http.StatusInternalServerError)
		return
	}
	jobs := extractJobs(doc, req.StartURL)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func main() {

	starturl := "https://www.pracuj.pl/praca/praca%20zdalna;wm,home-office?et=1%2C3%2C17" //praktykant-stażysta
	 // + 1,2,3,4,5, etc.

doc := getObj(starturl)
jobs :=  extractJobs(doc, starturl)		// to będzie się powtarzać

file, error := os.Create("posts.csv")
checkerror(error)
writer := csv.NewWriter(file)
for _, job := range jobs {
	writer.Write([]string{job.Title, job.Link})
}
writer.Flush()
//fmt.Println("Total pages:", findNumberPages(doc))

	http.HandleFunc("/jobs", jobsHandler)
	http.ListenAndServe(":8080", nil)
}