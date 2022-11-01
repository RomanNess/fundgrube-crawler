package crawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
	outlets, err := fetchOutlets(shop, mockedPostings)
	if err != nil {
		return nil, err
	}

	if envBool("LIMIT_OUTLETS") {
		outlets = outlets[0:5]
	}

	for _, outlet := range outlets {
		// api always returns same page if limit >= 100 is requested
		limit := 90
		offset := 0
		for true {
			postingsResponse, e := fetchSinglePageOfPostings(shop, &outlet, nil, limit, offset, mockedPostings)
			if e != nil {
				return nil, e
			}
			ret = append(ret, postingsResponse.Postings...)

			// api cannot request offset > 990; iterate over outlets or brands
			offset = offset + limit
			if !postingsResponse.HasMorePages || offset > 990 {
				break
			}
		}
	}
	log.Printf("Found %d Postings overall in %d outlets.", len(ret), len(outlets))
	return preparePostings(shop, ret), err
}

func preparePostings(shop Shop, postings []posting) []posting {
	for i, p := range postings {
		postings[i] = preparePosting(shop, p)
	}
	return postings
}

func preparePosting(shop Shop, posting posting) posting {
	posting.Shop = shop
	posting.ShopUrl = buildUrl(shop, &posting.Outlet, &posting.Brand, nil)
	posting.Price, _ = strconv.ParseFloat(posting.PriceString, 64)
	posting.PriceOld, _ = strconv.ParseFloat(posting.PriceOldString, 64)
	posting.PriceString = ""
	posting.PriceOldString = ""
	for i := range posting.Url {
		posting.Url[i] = fmt.Sprintf("%s?strip=yes&quality=75&backgroundsize=cover&x=640&y=640", posting.Url[i])
	}
	return posting
}

func fetchOutlets(shop Shop, mockedPostings bool) ([]outlet, error) {
	postingsResponse, err := fetchSinglePageOfPostings(shop, nil, nil, 1, 0, mockedPostings)
	if err != nil {
		return nil, err
	}
	log.Printf("Found %d Outlets", len(postingsResponse.Outlets))
	return postingsResponse.Outlets, err
}

func fetchSinglePageOfPostings(shop Shop, outlet *outlet, brand *brand, limit int, offset int, mockedPostings bool) (*postingsResponse, error) {
	urlString := buildUrl(shop, outlet, brand, &pageRequest{limit, offset})
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

	if outlet != nil {
		log.Printf("Found %d postings in %s.", len(postingResponse.Postings), outlet.Name)
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

	log.Printf("Querying: %s", url)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	responseBody := response.Body
	return responseBody, err
}

func buildUrl(shop Shop, outlet *outlet, brand *brand, pageRequest *pageRequest) string {
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
	if outlet != nil {
		q.Set("outletIds", strconv.Itoa(outlet.OutletId))
	}
	if brand != nil {
		q.Set("brands", url.QueryEscape(brand.Name))
	}

	u.RawQuery = q.Encode()
	return u.String()
}

func buildBaseUrl(shop Shop, isApiRequest bool) string {
	if shop == SATURN {
		if isApiRequest {
			return "https://www.saturn.de/de/data/fundgrube/api/postings?categorieIds=CAT_DE_SAT_786"
		}
		return "https://www.saturn.de/de/data/fundgrube?categorieIds=CAT_DE_SAT_786"
	}
	if shop == MM {
		if isApiRequest {
			return "https://www.mediamarkt.de/de/data/fundgrube/api/postings?categorieIds=CAT_DE_MM_626"
		}
		return "https://www.mediamarkt.de/de/data/fundgrube?categorieIds=CAT_DE_MM_626"
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
