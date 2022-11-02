package crawler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_preparePosting(t *testing.T) {
	type args struct {
		shop    Shop
		posting posting
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
			},
			posting{
				Brand:             brand{10, "Sony"},
				Outlet:            postingOutlet{100, "Outlet"},
				Price:             12.34,
				PriceString:       "",
				PriceOld:          24.68,
				PriceOldString:    "",
				DiscountInPercent: 50,
				Shop:              MM,
				ShopUrl:           "https://www.mediamarkt.de/de/data/fundgrube?brands=Sony&categorieIds=CAT_DE_MM_626&outletIds=100",
				Url: []string{
					"https://foo.bar?strip=yes&quality=75&backgroundsize=cover&x=640&y=640",
					"https://the.back?strip=yes&quality=75&backgroundsize=cover&x=640&y=640",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, preparePosting(tt.args.shop, tt.args.posting), "preparePosting(%v, %v)", tt.args.shop, tt.args.posting)
		})
	}
}

func Test_preparePostings(t *testing.T) {
	type args struct {
		shop     Shop
		postings []posting
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
			},
			[]posting{{
				Brand:             brand{10, "Sony"},
				Outlet:            postingOutlet{100, "Outlet"},
				Price:             12.34,
				PriceString:       "",
				PriceOld:          24.68,
				PriceOldString:    "",
				DiscountInPercent: 50,
				Shop:              MM,
				ShopUrl:           "https://www.mediamarkt.de/de/data/fundgrube?brands=Sony&categorieIds=CAT_DE_MM_626&outletIds=100",
				Url: []string{
					"https://foo.bar?strip=yes&quality=75&backgroundsize=cover&x=640&y=640",
					"https://the.back?strip=yes&quality=75&backgroundsize=cover&x=640&y=640",
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postings := preparePostings(tt.args.shop, tt.args.postings)
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
