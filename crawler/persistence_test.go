package crawler

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"os"
	"testing"
	"time"
)

type PersistenceSuite struct {
	suite.Suite
}

const PID_CHEF_PARTY = "ffd51e3a-01e6-40fc-a6e3-c241fbd88a7a"
const PID_NECRODANCER = "ffd23648-6353-4c18-93d5-78e0ac838da1"
const PID_ASUS = "ffba6620-c96e-43f3-9e1e-05f1bb3f0981"

func (suite *PersistenceSuite) SetupTest() {
	err := os.Setenv("MONGODB_DB", "fundgrube_test")
	if err != nil {
		panic(err)
	}
	clearAll()
	insertPostingsFromJson(err)
}

func insertPostingsFromJson(err error) {
	var postings []interface{}
	bytes, err := os.ReadFile("../_test/postings.json")
	if err != nil {
		panic(err)
	}

	if err = bson.UnmarshalExtJSON(bytes, true, &postings); err != nil {
		panic(err)
	}
	_, err = postingsCollection().InsertMany(context.TODO(), postings)
	if err != nil {
		panic(err)
	}
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(PersistenceSuite))
}

func (suite *PersistenceSuite) Test_connect() {
	postingsCollection()
}

func (suite *PersistenceSuite) Test_findOne_nonexistent() {
	posting := findOne("does-not-exist")
	assert.Nil(suite.T(), posting)
}

func (suite *PersistenceSuite) Test_findOne() {
	expectedPosting := posting{
		PostingId:         PID_CHEF_PARTY,
		PriceString:       "",
		PriceOldString:    "",
		Price:             10.0,
		PriceOld:          27.99,
		DiscountInPercent: 64,
		ShippingCost:      0,
		ShippingType:      "shipping",
		PimId:             1111111,
		Name:              "Instant Chef Party - [Nintendo Switch]",
		Url:               []string{"https://assets.mmsrg.com/is/166325/12975367df8e182e57044734f5165e190/c3/-/05154e6b51204fa699e88d114dba9b6d?strip=yes&quality=75&backgroundsize=cover&x=640&y=640"},
		Text:              "Neu, Verpackungsschaden / Folie kann beschädigt sein. OVP",
		Outlet:            postingOutlet{475, "Lübeck"},
		Brand:             brand{10312, "WILD RIVER"},
		Shop:              MM,
		ShopUrl:           "https://www.mediamarkt.de/de/data/fundgrube?brands=WILD%2BRIVER&categorieIds=CAT_DE_MM_626&outletIds=475",
		CreDat:            parseDate("2022-10-27T18:10:00.796Z"),
		ModDat:            parseDate("2022-10-31T21:47:16.898Z"),
	}

	assert.Equal(suite.T(), expectedPosting, *findOne(PID_CHEF_PARTY))
}

func (suite *PersistenceSuite) Test_findAll() {
	type args struct {
		q         query
		afterTime *time.Time
		limit     int64
		offset    int64
	}
	tests := []struct {
		name        string
		args        args
		expectedIds []string
	}{
		{
			"find all",
			args{query{}, nil, 100, 0},
			[]string{PID_CHEF_PARTY, PID_NECRODANCER, PID_ASUS},
		}, {
			"offset 1",
			args{query{}, nil, 100, 1},
			[]string{PID_NECRODANCER, PID_ASUS},
		}, {
			"limit 2",
			args{query{}, nil, 2, 0},
			[]string{PID_CHEF_PARTY, PID_NECRODANCER},
		}, {
			"regex search",
			args{query{NameRegex: sPtr("^.*nintendo.*$")}, nil, 100, 0},
			[]string{PID_CHEF_PARTY, PID_NECRODANCER},
		}, {
			"price min/max",
			args{query{PriceMin: fPtr(15), PriceMax: fPtr(30)}, nil, 100, 0},
			[]string{PID_NECRODANCER},
		}, {
			"discount min",
			args{query{DiscountMin: iPtr(60)}, nil, 100, 0},
			[]string{PID_CHEF_PARTY},
		}, {
			"brand regex",
			args{query{BrandRegex: sPtr("nin.*o")}, nil, 100, 0},
			[]string{PID_NECRODANCER},
		}, {
			"outletId",
			args{query{OutletId: iPtr(475)}, nil, 100, 0},
			[]string{PID_CHEF_PARTY, PID_NECRODANCER},
		}, {
			"after time",
			args{query{}, parseDate("2022-10-30T00:00:00Z"), 100, 0},
			[]string{PID_CHEF_PARTY, PID_ASUS},
		}, {
			"search by ids",
			args{query{Ids: []string{PID_NECRODANCER, PID_ASUS}}, nil, 100, 0},
			[]string{PID_NECRODANCER, PID_ASUS},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			postings := findAll(tt.args.q, tt.args.afterTime, tt.args.limit, tt.args.offset)
			var postingIds []string
			for _, p := range postings {
				postingIds = append(postingIds, p.PostingId)
			}
			assert.Equal(suite.T(), tt.expectedIds, postingIds)
		})
	}
}

func (suite *PersistenceSuite) Test_findAll_findNew() {
	foo := getExamplePosting("foo")
	saveAllNewOrUpdated([]posting{foo})

	postings := findAll(query{}, nil, 100, 3)
	assert.Equal(suite.T(), 1, len(postings))
	assertPostingsContainIgnoringDates(suite.T(), postings, foo)
}

func (suite *PersistenceSuite) Test_saveAll_updateName() {
	p := findOne(PID_CHEF_PARTY)

	p.Name = "New Name"
	inserted, updated, _ := saveAllNewOrUpdated([]posting{*p})

	assert.Equal(suite.T(), "New Name", findOne(p.PostingId).Name)
	assert.Equal(suite.T(), 0, inserted)
	assert.Equal(suite.T(), 1, updated)
}

func (suite *PersistenceSuite) Test_saveAll() {
	alreadySaved := getExamplePosting("alreadySaved")
	saveAllNewOrUpdated([]posting{alreadySaved})
	alreadySaved.Name = "New Name"
	notSavedYet := getExamplePosting("notSavedYet")

	inserted, updated, took := saveAllNewOrUpdated([]posting{alreadySaved, notSavedYet})

	all := findAll(query{}, nil, 100, 0)
	assertPostingsContainIgnoringDates(suite.T(), all, alreadySaved)
	assertPostingsContainIgnoringDates(suite.T(), all, notSavedYet)

	assert.Equal(suite.T(), 1, inserted)
	assert.Equal(suite.T(), 1, updated)
	assert.NotNil(suite.T(), took)
}

func (suite *PersistenceSuite) Test_saveOperation() {
	now := time.Now().UTC().Round(time.Millisecond)
	hash := getExampleHash()

	updateSearchOperation(getExampleQuery(), &now)
	assert.Equal(suite.T(), operation{hash, "description", getExampleQuery(), &now}, *findSearchOperation(hash))
}

func (suite *PersistenceSuite) Test_updateOperation() {
	now := time.Now().UTC().Round(time.Millisecond)
	hash := getExampleHash()
	updateSearchOperation(getExampleQuery(), &now)
	assert.Equal(suite.T(), operation{hash, "description", getExampleQuery(), &now}, *findSearchOperation(hash))

	now2 := now.AddDate(0, 0, 1)
	updateSearchOperation(getExampleQuery(), &now2)
	assert.Equal(suite.T(), operation{hash, "description", getExampleQuery(), &now2}, *findSearchOperation(hash))
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

func parseDate(dateString string) *time.Time {
	parse, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		panic(err)
	}
	return &parse
}

func ptr(s string) *string {
	return &s
}

func getExamplePosting(prefix string) posting {
	return posting{
		PostingId:         prefix + "-id",
		Name:              prefix,
		Text:              prefix + " text",
		Url:               []string{"http://" + prefix},
		Price:             1337.00,
		PriceOld:          1338.00,
		DiscountInPercent: 1,
		Outlet:            postingOutlet{42, "outlet"},
	}
}

func getExampleHash() string {
	return hashQuery(getExampleQuery())
}

func getExampleQuery() query {
	return query{NameRegex: sPtr("keyword"), Desc: "description"}
}
