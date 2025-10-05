package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
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
	ETypes   []int  `json:"et,omitempty"` // optional: multiple et codes, e.g. [1,3]
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

func extractJobs(doc *goquery.Document, baseURL string, isintern bool) []Job {
	var jobs []Job
	pages := findNumberPages(doc)
	log.Printf("Total pages: %d | start URL: %s | isIntern=%v", pages, baseURL, isintern)

	seen := make(map[string]bool)
	for i := 1; i <= pages; i++ {
		newURL := withPn(baseURL, i)
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

// withPn sets pn only when page > 1, preserving existing query params.
func withPn(base string, page int) string {
	u, err := url.Parse(base)
	if err != nil {
		log.Printf("withPn: parse error for %q: %v", base, err)
		return base
	}
	q := u.Query()
	if page > 1 {
		q.Set("pn", strconv.Itoa(page))
	} else {
		q.Del("pn")
	}
	u.RawQuery = q.Encode()
	return u.String()
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

func buildURL(query string, etCodes []int) string {
	u, _ := url.Parse("https://www.pracuj.pl/praca/" + query + ";kw")
	q := u.Query()
	if len(etCodes) > 0 {
		parts := make([]string, 0, len(etCodes))
		for _, c := range etCodes {
			parts = append(parts, strconv.Itoa(c))
		}
		// Results in et=1,3 (encoded as 1%2C3). Both are accepted by the site.
		q.Set("et", strings.Join(parts, ","))
	}
	u.RawQuery = q.Encode()
	out := u.String()
	log.Printf("buildURL: query=%q et=%v -> %s (rawQuery=%q)", query, etCodes, out, u.RawQuery)
	return out
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
	log.Printf("Request: query=%q isIntern=%v et=%v", req.Query, req.IsIntern, req.ETypes)

	// Back-compat: IsIntern implies et=1 if no explicit et list provided.
	et := req.ETypes
	if len(et) == 0 && req.IsIntern {
		et = []int{1}
	}

	URL := buildURL(req.Query, et)
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
