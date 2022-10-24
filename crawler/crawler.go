package crawler

import (
	"log"
	"os"
)

func Crawl(mockedPostings bool) error {
	postings, err := fetchPostings(500, mockedPostings)
	if err != nil {
		return err
	}

	query := getQuery()
	deals := findDeals(postings, query)

	presentDeals(deals)

	return nil
}

func presentDeals(deals []posting) {
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
