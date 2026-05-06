package model

import "testing"

func TestParseCoord(t *testing.T) {
	cases := []struct {
		in      string
		x, y    int
		wantErr bool
	}{
		{"A1", 0, 0, false},
		{"a1", 0, 0, false},
		{"B7", 1, 6, false},
		{"J10", 9, 9, false},
		{" J10 ", 9, 9, false},
		{"K1", 0, 0, true},
		{"A0", 0, 0, true},
		{"A11", 0, 0, true},
		{"A", 0, 0, true},
		{"", 0, 0, true},
	}
	for _, c := range cases {
		x, y, err := ParseCoord(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("ParseCoord(%q) esperava erro", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseCoord(%q) erro inesperado: %v", c.in, err)
			continue
		}
		if x != c.x || y != c.y {
			t.Errorf("ParseCoord(%q) = (%d,%d); quer (%d,%d)", c.in, x, y, c.x, c.y)
		}
	}
}

func TestFormatCoord(t *testing.T) {
	if got := FormatCoord(1, 6); got != "B7" {
		t.Errorf("FormatCoord(1,6) = %q; quer B7", got)
	}
}
