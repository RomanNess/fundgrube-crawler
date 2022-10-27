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
					Brand:  brand{10, "Sony"},
					Outlet: outlet{100, "Outlet"},
					Url:    []string{"https://foo.bar", "https://the.back"},
				},
			},
			posting{
				Brand:   brand{10, "Sony"},
				Outlet:  outlet{100, "Outlet"},
				Shop:    MM,
				ShopUrl: "https://www.mediamarkt.de/de/data/fundgrube?brands=Sony&categorieIds=CAT_DE_MM_626&outletIds=100",
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
				[]posting{posting{
					Brand:  brand{10, "Sony"},
					Outlet: outlet{100, "Outlet"},
					Url:    []string{"https://foo.bar", "https://the.back"},
				}},
			},
			[]posting{posting{
				Brand:   brand{10, "Sony"},
				Outlet:  outlet{100, "Outlet"},
				Shop:    MM,
				ShopUrl: "https://www.mediamarkt.de/de/data/fundgrube?brands=Sony&categorieIds=CAT_DE_MM_626&outletIds=100",
				Url: []string{
					"https://foo.bar?strip=yes&quality=75&backgroundsize=cover&x=640&y=640",
					"https://the.back?strip=yes&quality=75&backgroundsize=cover&x=640&y=640",
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preparePostings(tt.args.shop, tt.args.postings)
		})
	}
}
