package crawler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_preparePosting(t *testing.T) {
	type args struct {
		shop Shop
		p    posting
		c    category
	}
	tests := []struct {
		name string
		args args
		want posting
	}{
		{
			"initialize fields",
			args{
				MM,
				posting{
					Brand:             brand{10, "Sony"},
					Outlet:            postingOutlet{100, "Outlet"},
					PriceString:       "12.34",
					PriceOldString:    "24.68",
					DiscountInPercent: 50,
					Url:               []string{"https://foo.bar", "https://the.back"},
				},
				category{CategoryId: "CAT_ID", Name: "Category", Count: 1234},
			},
			posting{
				Brand:             brand{10, "Sony"},
				Outlet:            postingOutlet{100, "Outlet"},
				Category:          postingCategory{"CAT_ID", "Category"},
				Price:             12.34,
				PriceString:       "",
				PriceOld:          24.68,
				PriceOldString:    "",
				DiscountInPercent: 50,
				Shop:              MM,
				ShopUrl:           "https://www.mediamarkt.de/de/data/fundgrube?brands=Sony&categorieIds=CAT_ID&outletIds=100",
				Url: []string{
					"https://foo.bar?strip=yes&quality=75&backgroundsize=cover&x=640&y=640",
					"https://the.back?strip=yes&quality=75&backgroundsize=cover&x=640&y=640",
				},
				Active: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, preparePosting(tt.args.shop, tt.args.p, tt.args.c), "preparePosting(%v, %v, %v)", tt.args.shop, tt.args.p, tt.args.c)
		})
	}
}

func Test_preparePostings(t *testing.T) {
	type args struct {
		shop     Shop
		postings []posting
		c        category
	}
	tests := []struct {
		name string
		args args
		want []posting
	}{
		{
			"empty slice",
			args{
				MM,
				[]posting{},
				category{CategoryId: "CAT_ID", Name: "Category", Count: 1234},
			},
			[]posting{},
		},
		{
			"initialize fields",
			args{
				MM,
				[]posting{{
					Brand:             brand{10, "Sony"},
					Outlet:            postingOutlet{100, "Outlet"},
					PriceString:       "12.34",
					PriceOldString:    "24.68",
					DiscountInPercent: 50,
					Url:               []string{"https://foo.bar", "https://the.back"},
				}},
				category{CategoryId: "CAT_ID", Name: "Category", Count: 1234},
			},
			[]posting{{
				Brand:             brand{10, "Sony"},
				Outlet:            postingOutlet{100, "Outlet"},
				Category:          postingCategory{"CAT_ID", "Category"},
				Price:             12.34,
				PriceString:       "",
				PriceOld:          24.68,
				PriceOldString:    "",
				DiscountInPercent: 50,
				Shop:              MM,
				ShopUrl:           "https://www.mediamarkt.de/de/data/fundgrube?brands=Sony&categorieIds=CAT_ID&outletIds=100",
				Url: []string{
					"https://foo.bar?strip=yes&quality=75&backgroundsize=cover&x=640&y=640",
					"https://the.back?strip=yes&quality=75&backgroundsize=cover&x=640&y=640",
				},
				Active: true,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postings := preparePostings(tt.args.shop, tt.args.postings, tt.args.c)
			assert.Equal(t, tt.want, postings)
		})
	}
}

func Test_sliceOutlets(t *testing.T) {
	tests := []struct {
		name    string
		outlets []outlet
		wantRet [][]outlet
	}{
		{
			"three outlets",
			[]outlet{
				{1, "1", 500},
				{2, "2", 400},
				{3, "3", 300},
				{4, "4", 300},
				{5, "5", 400},
			},
			[][]outlet{
				{
					{1, "1", 500},
					{2, "2", 400},
				}, {
					{3, "3", 300},
					{4, "4", 300},
				}, {
					{5, "5", 400},
				},
			},
		}, {
			"huge outlet",
			[]outlet{
				{1, "1", 991},
				{2, "2", 400},
			},
			[][]outlet{
				{
					{1, "1", 991},
				}, {
					{2, "2", 400},
				},
			},
		}, {
			"no outlets",
			[]outlet{},
			[][]outlet{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantRet, sliceOutlets(tt.outlets), "sliceOutlets(%v)", tt.outlets)
		})
	}
}

func Test_buildUrl(t *testing.T) {
	type args struct {
		shop        Shop
		outlets     []outlet
		categories  []category
		brand       *brand
		pageRequest *pageRequest
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"set all",
			args{
				MM,
				[]outlet{
					{23, "Duisburg", 17},
					{24, "Düsseldorf", 18},
				},
				[]category{
					category{CategoryId: "CAT_ID", Name: "Category", Count: 1234},
					category{CategoryId: "CAT_ID_2", Name: "Category2", Count: 2345},
				},
				&brand{42, "A COOL BRAND"},
				&pageRequest{limit: 100, offset: 200},
			},
			"https://www.mediamarkt.de/de/data/fundgrube/api/postings?brands=A+COOL+BRAND&categorieIds=CAT_ID%2CCAT_ID_2&limit=100&offset=200&outletIds=23%2C24",
		}, {
			"set none results in shop link",
			args{
				MM,
				nil,
				nil,
				nil,
				nil,
			},
			"https://www.mediamarkt.de/de/data/fundgrube",
		}, {
			"empty outlets & categories",
			args{
				MM,
				[]outlet{},
				[]category{},
				nil,
				nil,
			},
			"https://www.mediamarkt.de/de/data/fundgrube",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, buildUrl(tt.args.shop, tt.args.outlets, tt.args.categories, tt.args.brand, tt.args.pageRequest), "buildUrl(%v, %v, %v, %v, %v)", tt.args.shop, tt.args.outlets, tt.args.categories, tt.args.brand, tt.args.pageRequest)
		})
	}
}

func Test_commaSeparatedOutletIds(t *testing.T) {
	type args struct {
		outlets []outlet
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"empty list",
			args{
				[]outlet{},
			},
			"",
		}, {
			"one outlet",
			args{
				[]outlet{{OutletId: 1}},
			},
			"1",
		}, {
			"two outlets",
			args{
				[]outlet{{OutletId: 1}, {OutletId: 2}},
			},
			"1,2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, commaSeparatedOutletIds(tt.args.outlets), "commaSeparatedOutletIds(%v)", tt.args.outlets)
		})
	}
}

func Test_filterCategories(t *testing.T) {
	type args struct {
		categories []category
		blacklist  []string
	}
	tests := []struct {
		name string
		args args
		want []category
	}{
		{
			"empty list",
			args{
				[]category{},
				[]string{},
			},
			[]category{},
		}, {
			"no blacklist",
			args{
				[]category{
					{"CAT_1", "Cat1", 1},
					{"CAT_2", "Cat2", 2},
				},
				[]string{},
			},
			[]category{
				{"CAT_1", "Cat1", 1},
				{"CAT_2", "Cat2", 2},
			},
		}, {
			"blacklist",
			args{
				[]category{
					{"CAT_1", "Cat1", 1},
					{"CAT_2", "Cat2", 2},
				},
				[]string{"CAT_1"},
			},
			[]category{
				{"CAT_2", "Cat2", 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, filterCategories(tt.args.categories, tt.args.blacklist), "filterCategories(%v, %v)", tt.args.categories, tt.args.blacklist)
		})
	}
}

func Test_toIds(t *testing.T) {
	type args struct {
		postingsForCategory []posting
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"empty list",
			args{[]posting{}},
			[]string{},
		}, {
			"list of one",
			args{[]posting{{PostingId: "fefe"}}},
			[]string{"fefe"},
		}, {
			"list of two",
			args{[]posting{{PostingId: "fefe"}, {PostingId: "dada"}}},
			[]string{"fefe", "dada"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, toIds(tt.args.postingsForCategory), "toIds(%v)", tt.args.postingsForCategory)
		})
	}
}
