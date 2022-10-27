package crawler

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"os"
	"testing"
	"time"
)

type PersistenceSuite struct {
	suite.Suite
}

func (suite *PersistenceSuite) SetupTest() {
	err := os.Setenv("MONGODB_DB", "fundgrube_test")
	if err != nil {
		log.Fatal(err)
	}
	clearAll()
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(PersistenceSuite))
}

func (suite *PersistenceSuite) Test_connect() {
	connectPostings()
}

func (suite *PersistenceSuite) Test_findOne_nonexistent() {
	posting := findOne("does-not-exist")
	assert.Nil(suite.T(), posting)
}

func (suite *PersistenceSuite) Test_findOne() {
	posting := getExamplePosting("foo")
	saveOne(posting)

	assertEqualPostingIgnoringDates(suite.T(), posting, *findOne("foo-id"))
}

func (suite *PersistenceSuite) Test_findAll_empty() {
	postings := findAll(getTime(), 100, 0)
	assert.Equal(suite.T(), []posting{}, postings)
}

func (suite *PersistenceSuite) Test_findAll() {
	foo := getExamplePosting("foo")
	saveOne(foo)
	bar := getExamplePosting("bar")
	saveOne(bar)

	postings := findAll(getTime(), 100, 0)
	assert.Equal(suite.T(), 2, len(postings))
	assertPostingsContainIgnoringDates(suite.T(), postings, foo)
	assertPostingsContainIgnoringDates(suite.T(), postings, bar)
}

func (suite *PersistenceSuite) Test_findAll_offset() {
	foo := getExamplePosting("foo")
	saveOne(foo)
	bar := getExamplePosting("bar")
	saveOne(bar)

	postings := findAll(getTime(), 1, 0)
	assert.Equal(suite.T(), 1, len(postings))
	assertPostingsContainIgnoringDates(suite.T(), postings, foo)

	postings = findAll(getTime(), 1, 1)
	assert.Equal(suite.T(), 1, len(postings))
	assertPostingsContainIgnoringDates(suite.T(), postings, bar)
}

func (suite *PersistenceSuite) Test_saveOne() {
	posting := getExamplePosting("foo")
	saveOne(posting)
	assertEqualPostingIgnoringDates(suite.T(), posting, *findOne(posting.PostingId))
}

func (suite *PersistenceSuite) Test_saveOne_updateName() {
	posting := getExamplePosting("foo")
	saveOne(posting)

	posting.Name = "New Name"
	saveOne(posting)

	assert.Equal(suite.T(), "New Name", findOne(posting.PostingId).Name)
}

func (suite *PersistenceSuite) Test_saveAll() {
	alreadySaved := getExamplePosting("alreadySaved")
	saveOne(alreadySaved)
	alreadySaved.Name = "New Name"
	notSavedYet := getExamplePosting("notSavedYet")

	saveAll([]posting{alreadySaved, notSavedYet})

	all := findAll(getTime(), 100, 0)
	assertPostingsContainIgnoringDates(suite.T(), all, alreadySaved)
	assertPostingsContainIgnoringDates(suite.T(), all, notSavedYet)
}

func (suite *PersistenceSuite) Test_saveOperation() {
	now := time.Now().UTC().Round(time.Millisecond)
	result := UpdateSearchOperation(&now)
	assert.Nil(suite.T(), result.Err())

	assert.Equal(suite.T(), operation{OP_SEARCH, &now}, *findSearchOperation())
}

func (suite *PersistenceSuite) Test_updateOperation() {
	now := time.Now().UTC().Round(time.Millisecond)
	UpdateSearchOperation(&now)
	assert.Equal(suite.T(), operation{OP_SEARCH, &now}, *findSearchOperation())

	now2 := now.AddDate(0, 0, 1)
	UpdateSearchOperation(&now2)
	assert.Equal(suite.T(), operation{OP_SEARCH, &now2}, *findSearchOperation())
}

func assertEqualPostingIgnoringDates(t *testing.T, expected posting, actual posting) {
	actual.CreDat = expected.CreDat
	actual.ModDat = expected.ModDat
	assert.Equal(t, expected, actual)
}

func assertPostingsContainIgnoringDates(t *testing.T, postings []posting, contained posting) bool {
	for i, p := range postings {
		p.CreDat = nil
		p.ModDat = nil
		postings[i] = p
	}
	return assert.Contains(t, postings, contained)
}

func getTime() *time.Time {
	yesterday := time.Now().AddDate(0, 0, -1)
	return &yesterday
}

func getExamplePosting(prefix string) posting {
	return posting{
		PostingId: prefix + "-id",
		Name:      prefix,
		Text:      prefix + " text",
		Url:       []string{"http://" + prefix},
		Price:     42.00,
		Outlet:    outlet{1337, "outlet"},
	}
}
