package crawler

import "strings"

func findDeals(postings *postings, query query) (deals []posting) {
	for _, posting := range postings.Postings {
		if isDeal(posting, query) {
			deals = append(deals, posting)
		}
	}
	return
}

func isDeal(posting posting, query query) bool {
	return strings.Contains(strings.ToLower(posting.Name), query.Keyword)
}
