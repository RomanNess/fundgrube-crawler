package crawler

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Shop string

const (
	SATURN Shop = "SATURN"
	MM          = "MM"
)

type pageRequest struct {
	limit  int
	offset int
}

func fetchPostings(shop Shop, mockedPostings bool) (ret []posting, err error) {
	categories, err := fetchCategories(shop, mockedPostings)
	if err != nil {
		return nil, err
	}

	for _, c := range categories {
		postingsForCategory, err2 := fetchPostingsForCategory(shop, mockedPostings, c)
		if err2 != nil {
			return nil, err2
		}
		ret = append(ret, postingsForCategory...)
	}
	return ret, nil
}

func fetchPostingsForCategory(shop Shop, mockedPostings bool, c category) ([]posting, error) {
	outlets, err := fetchOutlets(shop, c, mockedPostings)
	if err != nil {
		return nil, err
	}

	postingsForCategory := []posting{}

	if envBool("LIMIT_OUTLETS") && len(outlets) > 5 {
		outlets = outlets[0:5]
	}

	for _, outlets := range sliceOutlets(outlets) {
		// api always returns same page if limit >= 100 is requested
		limit := 90
		offset := 0
		for true {
			postingsResponse, err := fetchSinglePageOfPostings(shop, outlets, []category{c}, nil, limit, offset, mockedPostings)
			if err != nil {
				return nil, err
			}
			postingsForCategory = append(postingsForCategory, postingsResponse.Postings...)

			// api cannot request offset > 990; iterate over outlets or brands
			offset = offset + limit
			if !postingsResponse.HasMorePages || offset > 990 {
				break
			}
		}
	}
	return preparePostings(shop, postingsForCategory, c), nil
}

func sliceOutlets(outlets []outlet) [][]outlet {
	ret := [][]outlet{}

	var numberOfPostings = 0
	var outletSlice []outlet
	for i, o := range outlets {
		outletSlice = append(outletSlice, o)
		numberOfPostings = numberOfPostings + o.Count
		if i+1 < len(outlets) && numberOfPostings+outlets[i+1].Count > 990 {
			ret = append(ret, outletSlice)
			outletSlice = []outlet{}
			numberOfPostings = 0
		}
	}
	if len(outletSlice) > 0 {
		ret = append(ret, outletSlice)
	}
	return ret
}

func preparePostings(shop Shop, postings []posting, c category) []posting {
	for i, p := range postings {
		postings[i] = preparePosting(shop, p, c)
	}
	return postings
}

func preparePosting(shop Shop, posting posting, c category) posting {
	posting.Shop = shop
	posting.ShopUrl = buildUrl(shop, []outlet{{OutletId: posting.Outlet.OutletId}}, []category{c}, &posting.Brand, nil)
	posting.Category = c.toPostingCategory()
	posting.Price, _ = strconv.ParseFloat(posting.PriceString, 64)
	posting.PriceOld, _ = strconv.ParseFloat(posting.PriceOldString, 64)
	posting.PriceString = ""
	posting.PriceOldString = ""
	posting.Active = true
	for i := range posting.Url {
		posting.Url[i] = fmt.Sprintf("%s?strip=yes&quality=75&backgroundsize=cover&x=640&y=640", posting.Url[i])
	}
	return posting
}

func fetchCategories(shop Shop, mockedPostings bool) ([]category, error) {
	postingsResponse, err := fetchSinglePageOfPostings(shop, nil, nil, nil, 1, 0, mockedPostings)
	if err != nil {
		return nil, err
	}
	log.Infof("Discovered %d Categories for %s", len(postingsResponse.Categories), shop)
	return postingsResponse.Categories, err
}

func fetchOutlets(shop Shop, c category, mockedPostings bool) ([]outlet, error) {
	postingsResponse, err := fetchSinglePageOfPostings(shop, nil, []category{c}, nil, 1, 0, mockedPostings)
	if err != nil {
		return nil, err
	}
	log.Infof("Discovered %d Outlets for %s and category '%s'", len(postingsResponse.Outlets), shop, c.Name)
	return postingsResponse.Outlets, err
}

func fetchSinglePageOfPostings(shop Shop, outlets []outlet, categories []category, brand *brand, limit int, offset int, mockedPostings bool) (*postingsResponse, error) {
	urlString := buildUrl(shop, outlets, categories, brand, &pageRequest{limit, offset})
	responseBodyReader, err := getResponseBody(urlString, mockedPostings)
	if err != nil {
		return nil, err
	}

	if responseBodyReader == nil {
		return nil, errors.New("responseBody was nil")
	}

	defer responseBodyReader.Close()
	body, err := ioutil.ReadAll(responseBodyReader)
	if err != nil {
		return nil, err
	}

	postingResponse := postingsResponse{}
	err = json.Unmarshal(body, &postingResponse)
	if err != nil {
		return nil, err
	}

	if outlets != nil {
		log.Infof("Fetched %d postings in %d outlets with offset %d.", len(postingResponse.Postings), len(outlets), offset)
	}
	return &postingResponse, nil
}

func getResponseBody(url string, mockedResponse bool) (io.ReadCloser, error) {
	if mockedResponse {
		return getResponseBodyFromMock()
	}
	return getResponseBodyFromServer(url)
}

func getResponseBodyFromServer(url string) (io.ReadCloser, error) {
	client := http.Client{Timeout: 5 * time.Second}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36")

	log.Debugf("Querying: %s", url)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode > 200 {
		return nil, errors.New(fmt.Sprintf("Http Status %d on call of '%s'", response.StatusCode, url))
	}
	responseBody := response.Body
	return responseBody, err
}

func buildUrl(shop Shop, outlets []outlet, categories []category, brand *brand, pageRequest *pageRequest) string {
	isApiRequest := pageRequest != nil
	u, err := url.Parse(buildBaseUrl(shop, isApiRequest))
	if err != nil {
		panic(err)
	}

	q := u.Query()
	if isApiRequest {
		q.Set("limit", strconv.Itoa(pageRequest.limit))
		q.Set("offset", strconv.Itoa(pageRequest.offset))
	}
	if outlets != nil && len(outlets) > 0 {
		q.Set("outletIds", commaSeparatedOutletIds(outlets))
	}
	if categories != nil && len(categories) > 0 {
		q.Set("categorieIds", commaSeparatedCategoryIds(categories))
	}
	if brand != nil {
		q.Set("brands", brand.Name)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

func commaSeparatedCategoryIds(categories []category) string {
	categoryIds := []string{}
	for _, c := range categories {
		categoryIds = append(categoryIds, c.CategoryId)
	}
	return strings.Join(categoryIds, ",")
}

func commaSeparatedOutletIds(outlets []outlet) string {
	outletIds := []string{}
	for _, o := range outlets {
		outletIds = append(outletIds, strconv.Itoa(o.OutletId))
	}
	return strings.Join(outletIds, ",")
}

func buildBaseUrl(shop Shop, isApiRequest bool) string {
	if shop == SATURN {
		if isApiRequest {
			return "https://www.saturn.de/de/data/fundgrube/api/postings"
		}
		return "https://www.saturn.de/de/data/fundgrube"
	}
	if shop == MM {
		if isApiRequest {
			return "https://www.mediamarkt.de/de/data/fundgrube/api/postings"
		}
		return "https://www.mediamarkt.de/de/data/fundgrube"
	}

	panic(fmt.Sprintf("Unkown Shop %s", shop))
}

func getResponseBodyFromMock() (io.ReadCloser, error) {
	file, err := os.Open("mock/postingsResponse.json")
	if err != nil {
		return nil, err
	}

	return io.NopCloser(file), nil
}

func envBool(key string) bool {
	return os.Getenv(key) == "true"
}
