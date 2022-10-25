package crawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func fetchPostings(mockedPostings bool) (ret []posting, err error) {
	outlets, err := fetchOutlets(mockedPostings)
	if err != nil {
		return nil, err
	}

	for _, outlet := range outlets {
		// api always returns same page if limit >= 100 is requested
		limit := 90
		offset := 0
		for true {
			postingsResponse, e := fetchSinglePageOfPostings(&outlet, limit, offset, mockedPostings)
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
	return
}

func fetchOutlets(mockedPostings bool) ([]outlet, error) {
	postingsResponse, err := fetchSinglePageOfPostings(nil, 1, 0, mockedPostings)
	if err != nil {
		return nil, err
	}
	return postingsResponse.Outlets, err
}

func fetchSinglePageOfPostings(outlet *outlet, limit int, offset int, mockedPostings bool) (*postingsResponse, error) {
	url := buildUrl(outlet, limit, offset)

	responseBodyReader, err := getResponseBody(url, mockedPostings)
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

func buildUrl(outlet *outlet, size int, offset int) string {
	url := fmt.Sprintf("https://www.saturn.de/de/data/fundgrube/api/postings?limit=%d&offset=%d&categorieIds=CAT_DE_SAT_786&recentFilter=categories", size, offset)
	if outlet == nil {
		return url
	}
	return url + "&outletIds=" + strconv.Itoa(outlet.OutletId)
}

func getResponseBodyFromMock() (io.ReadCloser, error) {
	file, err := os.Open("mock/postingsResponse.json")
	if err != nil {
		return nil, err
	}

	return io.NopCloser(file), nil
}
