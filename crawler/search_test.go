package crawler

import "testing"

func Test_isDeal(t *testing.T) {
	type args struct {
		posting posting
		query   query
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"keyword matches case insensitive",
			args{
				posting: getPosting("Super cheap Sony Walkman SE"),
				query:   query{Keyword: "sony walkman"},
			},
			true,
		}, {
			"keyword not contained",
			args{
				posting: getPosting("Super cheap Sony Discman SE"),
				query:   query{Keyword: "sony walkman"},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDeal(tt.args.posting, tt.args.query); got != tt.want {
				t.Errorf("isDeal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getPosting(name string) posting {
	return posting{
		Name: name,
	}
}
