package crawler

import "strings"

func findDeals(postings []posting, query query) (deals []posting) {
	for _, posting := range postings {
		if isDeal(posting, query) {
			deals = append(deals, posting)
		}
	}
	return
}

func isDeal(posting posting, query query) bool {
	return strings.Contains(strings.ToLower(posting.Name), query.Keyword)
}
