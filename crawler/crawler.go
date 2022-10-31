package crawler

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"fundgrube-crawler/alert"
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
	var limit, offset int64 = 100, 0
	deals := []posting{}
	for true {
		postings := findAll(getLastSearchTime(query), limit, offset)
		log.Printf("Loaded %d postings from DB.", len(postings))

		offset = offset + limit
		if len(postings) == 0 {
			break
		}
		deals = append(deals, findDeals(postings, query)...)
	}
	if len(deals) > 0 {
		message := fmtDealsMessage(deals)
		err := alert.SendAlertMail(fmt.Sprintf("Found %d new deals.", len(deals)), message)
		if err != nil {
			log.Fatal(err)
		}
	}
	updateSearchOperation(query, now())
}

func getLastSearchTime(q query) *time.Time {
	md5Hex := hashQuery(q)

	op := findSearchOperation(md5Hex)
	if op == nil {
		return &time.Time{}
	}
	return op.Timestamp
}

func hashQuery(q query) string {
	jsonBytes, err := json.Marshal(q)
	if err != nil {
		log.Fatal(err)
	}
	md5Bytes := md5.Sum(jsonBytes)
	return hex.EncodeToString(md5Bytes[:])
}

func now() *time.Time {
	now := time.Now().UTC().Round(time.Millisecond)
	return &now
}

func fmtDealsMessage(deals []posting) string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("Found %d deals.\n\n", len(deals)))
	for _, deal := range deals {
		buffer.WriteString(deal.String() + "\n\n")
	}

	message := buffer.String()
	log.Println(message)
	return message
}

func getQuery() query {
	keyword := os.Getenv("SEARCH_KEYWORD")
	if keyword == "" {
		log.Println("SEARCH_KEYWORD not set! Default to 'example'.")
		keyword = "example"
	}
	return query{keyword}
}
