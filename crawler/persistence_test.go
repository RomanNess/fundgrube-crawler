package crawler

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"os"
	"testing"
)

type PersistenceSuite struct {
	suite.Suite
}

func (suite *PersistenceSuite) SetupTest() {
	clearAll()
	err := os.Setenv("MONGODB_DB", "fundgrube_test")
	if err != nil {
		log.Fatal(err)
	}
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(PersistenceSuite))
}

func (suite *PersistenceSuite) Test_connect() {
	connect()
}

func (suite *PersistenceSuite) Test_findOne_nonexistent() {
	posting := findOne("does-not-exist")
	assert.Nil(suite.T(), posting)
}

func (suite *PersistenceSuite) Test_findOne() {
	posting := getExamplePosting("foo")
	saveOne(posting)

	assert.Equal(suite.T(), posting, *findOne("foo-id"))
}

func (suite *PersistenceSuite) Test_findAll_empty() {
	postings := findAll()
	assert.Equal(suite.T(), []posting{}, postings)
}

func (suite *PersistenceSuite) Test_findAll() {
	foo := getExamplePosting("foo")
	saveOne(foo)
	bar := getExamplePosting("bar")
	saveOne(bar)

	postings := findAll()
	assert.Equal(suite.T(), 2, len(postings))
	assert.Contains(suite.T(), postings, foo)
	assert.Contains(suite.T(), postings, bar)
}

func (suite *PersistenceSuite) Test_saveOne() {
	posting := getExamplePosting("foo")
	saveOne(posting)
	assert.Equal(suite.T(), posting, *findOne(posting.PostingId))
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

	all := findAll()
	assert.Contains(suite.T(), all, alreadySaved)
	assert.Contains(suite.T(), all, notSavedYet)
}

func getExamplePosting(prefix string) posting {
	return posting{
		PostingId: prefix + "-id",
		Name:      prefix,
		Text:      prefix + " text",
		Url:       []string{"http://" + prefix},
		Price:     "42.00",
		Outlet:    outlet{1337, "outlet"},
	}
}
