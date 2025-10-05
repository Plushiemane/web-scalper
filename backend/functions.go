package main

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Job struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

type Request struct {
	Query    string `json:"query"`
	IsIntern bool   `json:"isintern"`
}

func init() {
	// Timestamp + file:line in logs
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func checkerror(err error) {
	if err != nil {
		log.Printf("ERROR: %v\n%s", err, debug.Stack())
	}
}

func extractJobs(doc *goquery.Document, url string, isintern bool) []Job {
	var jobs []Job
	pages := findNumberPages(doc)
	log.Printf("Total pages: %d | start URL: %s | isIntern=%v", pages, url, isintern)

	appendix := "?pn=" // default for no existing query params
	if isintern {
		appendix = "&pn=" // intern URLs already have ?et=1
	}

	seen := make(map[string]bool)
	for i := 1; i <= pages; i++ {
		var newURL string
		if i == 1 {
			newURL = url
		} else {
			newURL = url + appendix + strconv.Itoa(i)
		}

		log.Printf("Fetching page %d: %s", i, newURL)
		pageDoc := getObj(newURL)
		if pageDoc == nil {
			log.Printf("WARN: nil document for page %d: %s", i, newURL)
			continue
		}

		count := 0
		added := 0
		sel := pageDoc.Find("div.tiles_b18pwp01.core_po9665q")
		if sel.Length() == 0 {
			log.Printf("WARN: selectors matched 0 elements on page %d: %s", i, newURL)
		}
		sel.Each(func(_ int, s *goquery.Selection) {
			count++
			title := strings.TrimSpace(s.Find("h2.tiles_h1p4o5k6").Text())
			link, _ := s.Find("a.tiles_cnb3rfy.core_n194fgoq").Attr("href")
			if link != "" && !seen[link] {
				jobs = append(jobs, Job{Title: title, Link: link})
				seen[link] = true
				added++
			}
		})
		log.Printf("Page %d: matched=%d added=%d total=%d", i, count, added, len(jobs))
	}
	log.Printf("Extraction finished. Total jobs: %d", len(jobs))
	return jobs
}

func getObj(url string) *goquery.Document {
	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		checkerror(err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("HTTP %d for %s", resp.StatusCode, url)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		checkerror(err)
		return nil
	}

	log.Printf("Fetched %s in %s", url, time.Since(start))
	return doc
}

func findNumberPages(doc *goquery.Document) int {
	maxPage := doc.Find(`span[data-test="top-pagination-max-page-number"]`).Text()
	number, _ := strconv.Atoi(maxPage)
	log.Printf("Pagination: raw=%q parsed=%d", maxPage, number)
	return number
}

func buildURL(query string, isIntern bool) string {
	starturl := "https://www.pracuj.pl/praca/"
	URL := starturl + query + ";kw"
	if isIntern {
		URL += "?et=1"
	}
	log.Printf("buildURL: query=%q isIntern=%v -> %s", query, isIntern, URL)
	return URL
}

func jobsHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("PANIC in jobsHandler: %v\n%s", rec, debug.Stack())
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}()

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
		log.Printf("Bad request body: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	log.Printf("Request: query=%q isIntern=%v", req.Query, req.IsIntern)

	URL := buildURL(req.Query, req.IsIntern)
	doc := getObj(URL)
	log.Printf("Start URL fetched: %s docNil=%v", URL, doc == nil)
	if doc == nil {
		http.Error(w, "Failed to fetch starturl", http.StatusInternalServerError)
		return
	}

	jobs := extractJobs(doc, URL, req.IsIntern)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(jobs); err != nil {
		checkerror(err)
	}
}
