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
				query{"foobar"},
			},
			"82916e4f1ff6c73dd1564660c85b7332",
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
			wantRet: []query{{"sony.*walkman"}},
		}, {
			name:    "default queries",
			env:     "",
			wantRet: []query{{"example"}},
		}, {
			name:    "two queries",
			env:     "sony.*walkman;other[^.]",
			wantRet: []query{{"sony.*walkman"}, {"other[^.]"}},
		},
	}
	for _, tt := range tests {
		_ = os.Setenv("QUERY_REGEX", tt.env)
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantRet, getQueries(), "getQueries()")
		})
	}
}
