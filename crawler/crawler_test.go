package crawler

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_hashQuery(t *testing.T) {
	type args struct {
		q query
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"simple query",
			args{
				query{NameRegex: sPtr("foobar")},
			},
			"4baeec635d7f15efa872a24506fc2390",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, hashQuery(tt.args.q), "hashQuery(%v)", tt.args.q)
		})
	}
}

func Test_getQueries(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		wantRet []query
	}{
		{
			name:    "single query",
			env:     "sony.*walkman",
			wantRet: []query{{Desc: "sony.*walkman", NameRegex: sPtr("sony.*walkman")}},
		}, {
			name:    "default queries",
			env:     "",
			wantRet: []query{{NameRegex: sPtr("example")}},
		}, {
			name: "two queries",
			env:  "sony.*walkman;other[^.]",
			wantRet: []query{
				{Desc: "sony.*walkman", NameRegex: sPtr("sony.*walkman")},
				{Desc: "other[^.]", NameRegex: sPtr("other[^.]")},
			},
		},
	}
	for _, tt := range tests {
		_ = os.Unsetenv("SEARCH_REQUEST_YAML")
		_ = os.Setenv("QUERY_REGEX", tt.env)
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantRet, getQueries(), "getQueries()")
		})
	}
}

func sPtr(s string) *string {
	return &s
}

func iPtr(i int) *int {
	return &i
}

func fPtr(f float64) *float64 {
	return &f
}
