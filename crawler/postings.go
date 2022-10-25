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
	"time"
)

func fetchPostings(mockedPostings bool) (ret []posting, err error) {
	// TODO: api always returns same page if size >= 100 is requested
	size := 90
	offset := 0
	for true {
		postingsResponse, e := fetchSinglePageOfPostings(size, offset, mockedPostings)
		if e != nil {
			return nil, e
		}
		ret = append(ret, postingsResponse.Postings...)

		// TODO: api cannot request offset > 990; iterate over outlets or brands
		offset = offset + size
		if !postingsResponse.HasMorePages || offset > 990 {
			return
		}
	}
	return
}

func fetchSinglePageOfPostings(limit int, offset int, mockedPostings bool) (*postingsResponse, error) {
	responseBodyReader, err := getResponseBody(limit, offset, mockedPostings)
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

	log.Printf("Found %d postings.", len(postingResponse.Postings))
	return &postingResponse, nil
}

func getResponseBody(limit int, offset int, mockedResponse bool) (io.ReadCloser, error) {
	if mockedResponse {
		return getResponseBodyFromMock()
	}
	return getResponseBodyFromServer(limit, offset)
}

func getResponseBodyFromServer(size int, offset int) (io.ReadCloser, error) {
	url := fmt.Sprintf("https://www.saturn.de/de/data/fundgrube/api/postings?limit=%d&offset=%d&categorieIds=CAT_DE_SAT_786&recentFilter=categories", size, offset)

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

func getResponseBodyFromMock() (io.ReadCloser, error) {
	file, err := os.Open("mock/postingsResponse.json")
	if err != nil {
		return nil, err
	}

	return io.NopCloser(file), nil
}
