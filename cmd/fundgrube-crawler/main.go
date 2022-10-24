package main

import (
	"fundgrube-crawler/crawler"
	"log"
)

func main() {
	err := crawler.Crawl(false)
	if err != nil {
		log.Fatal(err)
	}
}
