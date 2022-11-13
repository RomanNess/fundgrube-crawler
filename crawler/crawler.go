package crawler

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"fundgrube-crawler/alert"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"time"
)

func CrawlPostings(mockedPostings bool) error {
	for _, shop := range []Shop{SATURN, MM} {
		postings, err := fetchPostings(shop, mockedPostings)
		if err != nil {
			return err // TODO: log error and continue?
		}
		inserted, updated, took := SaveAllNewOrUpdated(postings)

		var ids []string
		for _, p := range postings {
			ids = append(ids, p.PostingId)
		}
		inactiveCount := SetRemainingPostingInactive(shop, ids)
		log.Infof("Refreshed %d Postings for %s. inserted: %d, updated: %d, inactive: %d, took: %fs", len(postings), shop, inserted, updated, inactiveCount, took.Seconds())

	}
	return nil
}

func SearchDeals() {
	queries := getQueries()
	for _, query := range queries {
		searchDealsForSingleQuery(query)
	}
}

func searchDealsForSingleQuery(query query) {
	var limit, offset int64 = 100, 0
	deals := []posting{}
	for true {
		postings := FindAll(query, getLastSearchTime(query), limit, offset)
		log.Infof("Found %d deals for query '%s'.", len(postings), query.Desc)
		deals = append(deals, postings...)

		if len(postings) < int(limit) {
			break
		}
		offset = offset + limit
	}
	if len(deals) > 0 {
		message := fmtDealsMessage(query, deals)
		err := alert.SendAlertMail(formatSubject(query, deals), message)
		if err != nil {
			log.Fatalf("Could not send deals via mail: %s", err)
		}
	}
	updateSearchOperation(query, now())
}

func formatSubject(q query, deals []posting) string {
	if len(deals) == 0 {
		return fmt.Sprintf("Found no deals for query '%s'. 😿", q.Desc)
	}
	deal := deals[0]
	return fmt.Sprintf("Query '%s' matched by %s for %.2f€ in %s (%d deal(s) overall)", q.Desc, deal.Name, deal.Price, deal.Outlet.Name, len(deals))
}

func getLastSearchTime(q query) *time.Time {
	if envBool("FIND_ALL") {
		return nil
	}

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
		panic(err)
	}
	md5Bytes := md5.Sum(jsonBytes)
	return hex.EncodeToString(md5Bytes[:])
}

func now() *time.Time {
	now := time.Now().UTC().Round(time.Millisecond)
	return &now
}

func fmtDealsMessage(q query, deals []posting) string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("Query '%s' matched by %d new deals.\n\n", q.Desc, len(deals)))
	for _, deal := range deals {
		buffer.WriteString(deal.String() + "\n\n")
	}

	message := buffer.String()
	log.Infoln(message)
	return message
}

func getQueries() (ret []query) {
	yamlPath := os.Getenv("SEARCH_REQUEST_YAML")
	if yamlPath != "" {
		return getQueriesFromYaml(yamlPath)
	}

	searchRegexes := os.Getenv("QUERY_REGEX")
	if searchRegexes != "" {
		return getQueriesFromEnv(searchRegexes)
	}

	log.Warnln("QUERY_REGEX not set! Default to 'example'.")
	regex := "example"
	return []query{{NameRegex: &regex}}
}

func getQueriesFromYaml(yamlPath string) []query {
	yamlBytes, err := os.ReadFile(yamlPath)
	if err != nil {
		panic(err)
	}
	searchRequest := queries{}
	err = yaml.Unmarshal(yamlBytes, &searchRequest)
	if err != nil {
		panic(err)
	}
	return searchRequest.Queries
}

func getQueriesFromEnv(searchRegexes string) (ret []query) {
	split := strings.Split(searchRegexes, ";")
	for _, regex := range split {
		regexCopy := regex // don't return a pointer on loop variable :)
		ret = append(ret, query{NameRegex: &regexCopy, Desc: regex})
	}
	return
}
