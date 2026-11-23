package teamgames

import (
	"database/sql"
	"testing"
)

func nullID(id int64) sql.NullInt64 {
	return sql.NullInt64{Int64: id, Valid: true}
}

func TestResolveMajorityCountryRegion(t *testing.T) {
	cases := []struct {
		name        string
		countries   map[int64]int
		regions     map[int64]int
		wantCountry sql.NullInt64
		wantRegion  sql.NullInt64
	}{
		{
			name:        "country majority wins",
			countries:   map[int64]int{1: 3, 2: 2},
			regions:     map[int64]int{10: 5},
			wantCountry: nullID(1),
		},
		{
			name:       "region majority wins when no country majority",
			countries:  map[int64]int{1: 2, 2: 2, 3: 1},
			regions:    map[int64]int{10: 4, 11: 1},
			wantRegion: nullID(10),
		},
		{
			name:      "no majority leaves both null",
			countries: map[int64]int{1: 2, 2: 2, 3: 2},
			regions:   map[int64]int{10: 3, 11: 3},
		},
		{
			name:      "exactly half is not a majority",
			countries: map[int64]int{1: 3, 2: 3},
			regions:   map[int64]int{10: 3, 11: 3},
		},
		{
			name:      "empty roster",
			countries: map[int64]int{},
			regions:   map[int64]int{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			total := 0
			for _, c := range tc.countries {
				total += c
			}
			country, region := resolveMajorityCountryRegion(tc.countries, tc.regions, total)
			if country != tc.wantCountry || region != tc.wantRegion {
				t.Fatalf("got country=%v region=%v, want country=%v region=%v",
					country, region, tc.wantCountry, tc.wantRegion)
			}
		})
	}
}
