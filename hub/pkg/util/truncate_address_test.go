package util

import "testing"

func TestTruncateAddress(t *testing.T) {
	cases := []struct {
		name     string
		address  string
		maxParts int
		want     string
	}{
		{
			name:     "nominatim address longer than maxParts",
			address:  "Piano Piano, 5, Rotušės a., Senamiestis, Centro seniūnija, Kaunas, Kauno miesto savivaldybė, Kauno apskritis, 44279, Lietuva",
			maxParts: 3,
			want:     "Piano Piano, 5, Rotušės a.",
		},
		{
			name:     "address shorter than maxParts is returned as is",
			address:  "Rotušės a., Kaunas",
			maxParts: 3,
			want:     "Rotušės a., Kaunas",
		},
		{
			name:     "address with exactly maxParts segments is returned as is",
			address:  "Piano Piano, 5, Rotušės a.",
			maxParts: 3,
			want:     "Piano Piano, 5, Rotušės a.",
		},
		{
			name:     "maxParts zero returns empty string",
			address:  "Piano Piano, 5, Rotušės a.",
			maxParts: 0,
			want:     "",
		},
		{
			name:     "extra whitespace around segments is trimmed",
			address:  "Piano Piano ,  5 , Rotušės a. ,Senamiestis",
			maxParts: 3,
			want:     "Piano Piano, 5, Rotušės a.",
		},
		{
			name:     "single segment address",
			address:  "Kaunas",
			maxParts: 3,
			want:     "Kaunas",
		},
		{
			name:     "empty address",
			address:  "",
			maxParts: 3,
			want:     "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := TruncateAddress(c.address, c.maxParts)
			if got != c.want {
				t.Errorf("TruncateAddress(%q, %d) = %q, want %q", c.address, c.maxParts, got, c.want)
			}
		})
	}
}
