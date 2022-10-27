package crawler

import (
	"log"
	"os"
	"time"
)

func CrawlPostings(mockedPostings bool) error {
	for _, shop := range []Shop{SATURN, MM} {
		postings, err := fetchPostings(shop, mockedPostings)
		if err != nil {
			return err
		}
		saveAll(postings)
	}
	return nil
}

func SearchDeals() {
	query := getQuery()
	var limit int64 = 100
	var offset int64 = 0
	deals := []posting{}
	for true {
		postings := findAll(getLastSearchTime(), limit, offset)
		log.Printf("Loaded %d postings from DB.", len(postings))

		offset = offset + limit
		if len(postings) == 0 {
			break
		}
		deals = append(deals, findDeals(postings, query)...)
	}
	presentDeals(deals)
	UpdateSearchOperation(now())
}

func getLastSearchTime() *time.Time {
	op := findSearchOperation()
	if op == nil {
		return &time.Time{}
	}
	return op.Timestamp
}

func now() *time.Time {
	now := time.Now().UTC().Round(time.Millisecond)
	return &now
}

func presentDeals(deals []posting) {
	log.Printf("Found %d deals.", len(deals))
	for _, deal := range deals {
		log.Println(deal)
	}
}

func getQuery() query {
	keyword := os.Getenv("SEARCH_KEYWORD")
	if keyword == "" {
		log.Println("SEARCH_KEYWORD not set! Default to 'example'.")
		keyword = "example"
	}
	return query{keyword}
}
