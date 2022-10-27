package main

import (
	"fundgrube-crawler/crawler"
	"log"
	"os"
)

func main() {
	if envBool("LOG_TO_FILE") {
		logToFile()
	}

	if !envBool("SKIP_CRAWLING") {
		err := crawler.CrawlPostings(envBool("MOCKED_POSTINGS"))
		if err != nil {
			log.Fatal(err)
		}
	}

	crawler.SearchDeals()
}

func logToFile() {
	file, err := os.OpenFile("/tmp/fundgrube.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Logging into '/tmp/fundgrube.txt'")
	log.SetOutput(file)
}

func envBool(key string) bool {
	return os.Getenv(key) == "true"
}
