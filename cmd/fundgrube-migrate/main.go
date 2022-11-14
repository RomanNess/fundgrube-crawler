package main

import (
	"fundgrube-crawler/crawler"
)

func main() {
	migrate()
}

func migrate() {
	// migrate schema
	crawler.Migrate(`{"outlet.outletid": {"$exists": 1}}`, `{"$rename": {"outlet.outletid": "outlet.id"}}`)
	crawler.Migrate(`{"brand.brandid": {"$exists": 1}}`, `{"$rename": {"brand.brandid": "brand.id"}}`)

	// clean up after bug
	crawler.CleanUp(`{"cre_dat": {"$eq": null}}`)
}
