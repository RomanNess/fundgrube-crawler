package main

import (
	"fundgrube-crawler/crawler"
	"log"
	"os"
)

func main() {
	if !envBool("SKIP_CRAWLING") {
		err := crawler.CrawlPostings(envBool("MOCKED_POSTINGS"))
		if err != nil {
			log.Fatal(err)
		}
	}

	crawler.SearchDeals()
}

func envBool(key string) bool {
	return os.Getenv(key) == "true"
}
