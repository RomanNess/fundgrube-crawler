package crawler

import (
	"github.com/stretchr/testify/assert"
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

func sPtr(s string) *string {
	return &s
}

func iPtr(i int) *int {
	return &i
}

func fPtr(f float64) *float64 {
	return &f
}
