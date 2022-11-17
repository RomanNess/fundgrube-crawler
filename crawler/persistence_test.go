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

const PID_NHL = "2822b32a-1057-4b21-ad2d-8d297a88d00c"
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
	posting := FindOne("does-not-exist")
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
		Outlet:            postingOutlet{111, "Lübeck"},
		Category:          postingCategory{"CAT_DE_SAT_786", "Gaming + VR"},
		Brand:             brand{10312, "WILD RIVER"},
		Shop:              MM,
		ShopUrl:           "https://www.mediamarkt.de/de/data/fundgrube?brands=WILD%2BRIVER&categorieIds=CAT_DE_MM_626&outletIds=475",
		CreDat:            parseDate("2022-10-27T18:10:00.796Z"),
		ModDat:            parseDate("2022-10-31T21:47:16.898Z"),
		Active:            true,
	}

	assert.Equal(suite.T(), expectedPosting, *FindOne(PID_CHEF_PARTY))
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
			"find all; find inactive",
			args{query{FindInactive: true}, nil, 100, 0},
			[]string{PID_NHL, PID_CHEF_PARTY, PID_NECRODANCER, PID_ASUS},
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
			"price min",
			args{query{PriceMin: fPtr(15)}, nil, 100, 0},
			[]string{PID_NECRODANCER, PID_ASUS},
		}, {
			"price max",
			args{query{PriceMax: fPtr(30)}, nil, 100, 0},
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
			args{query{OutletId: iPtr(111)}, nil, 100, 0},
			[]string{PID_CHEF_PARTY},
		}, {
			"after time",
			args{query{}, parseDate("2022-10-30T00:00:00Z"), 100, 0},
			[]string{PID_CHEF_PARTY, PID_ASUS},
		}, {
			"search by ids",
			args{query{Ids: []string{PID_NECRODANCER, PID_ASUS}}, nil, 100, 0},
			[]string{PID_NECRODANCER, PID_ASUS},
		}, {
			"active only",
			args{query{FindInactive: false}, nil, 100, 0},
			[]string{PID_CHEF_PARTY, PID_NECRODANCER, PID_ASUS},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			postings := FindAll(tt.args.q, tt.args.afterTime, tt.args.limit, tt.args.offset)
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
	SaveAllNewOrUpdated([]posting{foo})

	postings := FindAll(query{}, nil, 100, 3)
	assert.Equal(suite.T(), 1, len(postings))
	assertPostingsContainIgnoringDates(suite.T(), postings, foo)
}

func (suite *PersistenceSuite) Test_saveAll_updateName() {
	p := FindOne(PID_CHEF_PARTY)

	p.Name = "New Name"
	stats := SaveAllNewOrUpdated([]posting{*p})

	assert.Equal(suite.T(), "New Name", FindOne(p.PostingId).Name)
	assert.Equal(suite.T(), 0, stats.Inserted)
	assert.Equal(suite.T(), 1, stats.Updated)
}

func (suite *PersistenceSuite) Test_saveAll() {
	alreadySaved := getExamplePosting("alreadySaved")
	SaveAllNewOrUpdated([]posting{alreadySaved})
	alreadySaved.Name = "New Name"
	notSavedYet := getExamplePosting("notSavedYet")

	stats := SaveAllNewOrUpdated([]posting{alreadySaved, notSavedYet})

	all := FindAll(query{}, nil, 100, 0)
	assertPostingsContainIgnoringDates(suite.T(), all, alreadySaved)
	assertPostingsContainIgnoringDates(suite.T(), all, notSavedYet)

	assert.Equal(suite.T(), 1, stats.Inserted)
	assert.Equal(suite.T(), 1, stats.Updated)
}

func (suite *PersistenceSuite) Test_insertOrUpdateAll_insertNew() {
	insertedCount, updatedCount := insertOrUpdateAll([]posting{getExamplePosting("foo")})
	assert.Equal(suite.T(), 1, insertedCount)
	assert.Equal(suite.T(), 0, updatedCount)
}

func (suite *PersistenceSuite) Test_insertOrUpdateAll_updateExisting() {
	insertedCount, updatedCount := insertOrUpdateAll([]posting{*FindOne(PID_CHEF_PARTY)})
	assert.Equal(suite.T(), 0, insertedCount)
	assert.Equal(suite.T(), 1, updatedCount)
}

func (suite *PersistenceSuite) Test_SetRemainingPostingInactive() {
	assert.Equal(suite.T(), true, FindOne(PID_ASUS).Active)
	assert.Equal(suite.T(), true, FindOne(PID_CHEF_PARTY).Active)
	assert.Equal(suite.T(), true, FindOne(PID_NECRODANCER).Active)
	SetRemainingPostingInactive(MM, category{"CAT_DE_SAT_786", "Cat1", 1}, []outlet{outl(111), outl(222)}, []string{PID_CHEF_PARTY})
	assert.Equal(suite.T(), true, FindOne(PID_ASUS).Active) // saturn
	assert.Equal(suite.T(), true, FindOne(PID_CHEF_PARTY).Active)
	assert.Equal(suite.T(), false, FindOne(PID_NECRODANCER).Active)
}

func (suite *PersistenceSuite) Test_SetRemainingPostingInactive_noActiveInCategoryAndOutlet() {
	assert.Equal(suite.T(), true, FindOne(PID_ASUS).Active)
	assert.Equal(suite.T(), true, FindOne(PID_CHEF_PARTY).Active)
	assert.Equal(suite.T(), true, FindOne(PID_NECRODANCER).Active)
	SetRemainingPostingInactive(MM, category{"CAT_DE_SAT_786", "Cat1", 1}, []outlet{outl(111)}, []string{})
	assert.Equal(suite.T(), true, FindOne(PID_ASUS).Active)        // saturn
	assert.Equal(suite.T(), false, FindOne(PID_CHEF_PARTY).Active) // outlet 111
	assert.Equal(suite.T(), true, FindOne(PID_NECRODANCER).Active)
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

func assertPostingsContainIgnoringDates(t *testing.T, postings []posting, contained posting) bool {
	postingsWithoutDates := []posting{}
	for _, p := range postings {
		assert.NotNil(t, p.CreDat)
		assert.NotNil(t, p.ModDat)
		p.CreDat = nil
		p.ModDat = nil
		postingsWithoutDates = append(postingsWithoutDates, p)
	}
	return assert.Contains(t, postingsWithoutDates, contained)
}

func parseDate(dateString string) *time.Time {
	parse, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		panic(err)
	}
	return &parse
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
		Active:            true,
	}
}

func getExampleHash() string {
	return hashQuery(getExampleQuery())
}

func getExampleQuery() query {
	return query{NameRegex: sPtr("keyword"), Desc: "description"}
}

func outl(id int) outlet {
	return outlet{OutletId: id, Name: "Outlet" + string(rune(id)), Count: 100 + id}
}
