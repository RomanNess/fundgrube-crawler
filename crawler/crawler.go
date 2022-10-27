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
	searchTime := getLastSearchTime()
	var limit int64 = 100
	var offset int64 = 0
	deals := []posting{}
	for true {
		postings := findAll(searchTime, limit, offset)
		log.Printf("Loaded %d postings from DB.", len(postings))

		offset = offset + limit
		if len(postings) == 0 {
			break
		}
		deals = append(deals, findDeals(postings, query)...)
	}
	presentDeals(deals)
}

func getLastSearchTime() time.Time {
	return time.Now().AddDate(0, 0, -1) // TODO: find things added since last search
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
