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
				query{Regex: sPtr("foobar")},
			},
			"297246b4db7859ed8f6c3ebb257b48e1",
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
			wantRet: []query{{Regex: sPtr("sony.*walkman")}},
		}, {
			name:    "default queries",
			env:     "",
			wantRet: []query{{Regex: sPtr("example")}},
		}, {
			name:    "two queries",
			env:     "sony.*walkman;other[^.]",
			wantRet: []query{{Regex: sPtr("sony.*walkman")}, {Regex: sPtr("other[^.]")}},
		},
	}
	for _, tt := range tests {
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
