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
)

func fetchPostings(pageSize int, mockedPostings bool) (*postings, error) {
	responseBody, err := getResponseBody(pageSize, mockedPostings)
	if err != nil {
		return nil, err
	}

	if responseBody == nil {
		return nil, errors.New("responseBody was nil")
	}

	defer responseBody.Close()
	body, err := ioutil.ReadAll(responseBody)
	if err != nil {
		return nil, err
	}

	postingResponse := postings{}
	err = json.Unmarshal(body, &postingResponse)
	if err != nil {
		return nil, err
	}

	log.Printf("Found %d postings.", len(postingResponse.Postings))
	return &postingResponse, nil
}

func getResponseBody(pageSize int, mocked bool) (io.ReadCloser, error) {
	if mocked {
		return getResponseBodyFromMock()
	}
	return getResponseBodyFromServer(pageSize)
}

// TODO: max supported limit is 100; add pagination
func getResponseBodyFromServer(pageSize int) (io.ReadCloser, error) {
	url := fmt.Sprintf("https://www.saturn.de/de/data/fundgrube/api/postings?limit=%d&offset=0&categorieIds=CAT_DE_SAT_786&recentFilter=categories", pageSize)

	client := http.Client{}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

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
