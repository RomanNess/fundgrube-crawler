package main

import (
	"fmt"
	"fundgrube-crawler/alert"
	"fundgrube-crawler/crawler"
	"log"
	"os"
	"time"
)

var LOG_FILE = fmt.Sprintf("/tmp/fundgrube-%s.txt", time.Now().Format("2006-01-02T15-04-05"))

func main() {
	if envBool("LOG_TO_FILE") {
		logToFile()
		defer mailAlertOnPanic()
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
	file, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Logging into '%s'", LOG_FILE)
	log.SetOutput(file)
}

func mailAlertOnPanic() {
	if r := recover(); r != nil {
		subject := fmt.Sprint("💥Panic occurred : ", r.(string))
		contentBytes := getContentBytes()
		err := alert.SendAlertMailBytes(subject, contentBytes)
		if err != nil {
			panic(err)
		}
		log.Println("💥Panic occurred. Send alert mail.", r)
	}
}

func getContentBytes() []byte {
	if envBool("LOG_TO_FILE") {
		contentBytes := []byte("\n\nLogs:\n\n")
		logBytes, err := os.ReadFile(LOG_FILE)
		if err != nil {
			panic(err)
		}
		return append(contentBytes, logBytes...)
	}
	return nil
}

func envBool(key string) bool {
	return os.Getenv(key) == "true"
}
