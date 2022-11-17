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
	"time"
)

var CONFIG ConfigFile

func RefreshAllPostings(mockedPostings bool) error {
	stats := CrawlerStats{}
	for _, shop := range []Shop{SATURN, MM} {
		categories, err := fetchCategories(shop, mockedPostings)
		if err != nil {
			return err
		}

		categories = filterCategories(categories, CONFIG.GlobalConfig.BlacklistedCategories)

		for _, c := range categories {
			categoryStats, err := RefreshPostingsForCategory(shop, mockedPostings, c)
			if err != nil {
				return err
			}
			stats.add(categoryStats)
		}
	}

	log.Infof("Refreshed postings. %s", stats.String())
	return nil
}

func RefreshOnlyNewPostings() error {
	log.Info("Fetching only new Postings.")
	stats := CrawlerStats{}
	for _, shop := range []Shop{SATURN, MM} {
		shopStats, err := refreshOnlyNewPostingsForShop(shop)
		if err != nil {
			return err
		}
		stats.add(shopStats)
	}
	return nil
}

func SearchDeals() {
	for _, query := range CONFIG.Queries {
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

func GetConfigFromFile(yamlPath string) ConfigFile {
	yamlBytes, err := os.ReadFile(yamlPath)
	if err != nil {
		panic(err)
	}
	cf := ConfigFile{}
	err = yaml.Unmarshal(yamlBytes, &cf)
	if err != nil {
		panic(err)
	}
	return cf
}
