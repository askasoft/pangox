package xargs

import (
	"strings"
	"testing"
)

func TestIntrgContains(t *testing.T) {
	tests := []struct {
		rng    intrg
		n      int64
		expect bool
	}{
		{intrg{nil, int64(10)}, 5, true},
		{intrg{nil, int64(10)}, 15, false},
		{intrg{int64(5), nil}, 6, true},
		{intrg{int64(5), nil}, 3, false},
		{intrg{int64(3), int64(7)}, 5, true},
		{intrg{int64(3), int64(7)}, 8, false},
	}

	for _, tt := range tests {
		got := tt.rng.Contains(tt.n)
		if got != tt.expect {
			t.Errorf("intrg(%v).Contains(%d) = %v, expected %v", tt.rng, tt.n, got, tt.expect)
		}
	}
}

func TestIntrgsContains(t *testing.T) {
	rs := intrgs{
		{int64(1), int64(5)},
		{int64(10), int64(15)},
	}
	if !rs.Contains(3) {
		t.Errorf("Expected Contains(3) = true")
	}
	if rs.Contains(6) {
		t.Errorf("Expected Contains(6) = false")
	}
}

func TestParseIntegers(t *testing.T) {
	tests := []struct {
		input   string
		wanterr bool
		check   func(Integers) bool
	}{
		{
			input: "-1 2 3",
			check: func(ns Integers) bool {
				return ns.Contains(-1) && ns.Contains(2) && ns.Contains(3)
			},
		},
		{
			input: "-1~3",
			check: func(ns Integers) bool {
				return ns.Contains(-1) && ns.Contains(0) && ns.Contains(1) && ns.Contains(2) && ns.Contains(3) && !ns.Contains(4)
			},
		},
		{
			input: "~3",
			check: func(ns Integers) bool {
				return ns.Contains(1) && ns.Contains(3) && !ns.Contains(4)
			},
		},
		{
			input: "5~",
			check: func(ns Integers) bool {
				return ns.Contains(5) && ns.Contains(1000)
			},
		},
		{
			input: "10~10",
			check: func(ns Integers) bool {
				return ns.Contains(10) && !ns.Contains(9)
			},
		},
		{
			input: "20~10",
			check: func(ns Integers) bool {
				return ns.Contains(10) && ns.Contains(20)
			},
		},
		{
			input: " ",
			check: func(ns Integers) bool {
				return len(ns.ints) == 0 && len(ns.ranges) == 0
			},
		},
		{
			input:   "1+a",
			wanterr: true,
		},
		{
			input:   "1 ~1+1",
			wanterr: true,
		},
		{
			input:   "1 2+a~",
			wanterr: true,
		},
	}

	for _, tt := range tests {
		ns, err := ParseIntegers(tt.input)
		if tt.wanterr {
			if err == nil {
				t.Errorf("ParseIntegers(%q) = %q, but want error", tt.input, ns.String())
			}
			continue
		}
		if !tt.check(ns) {
			t.Errorf("ParseIntegers(%q) failed", tt.input)
		}
	}
}

func TestIntegersString(t *testing.T) {
	input := "1 3 5~10 ~2 20~"
	ns, _ := ParseIntegers(input)
	expectedParts := []string{"1", "3", "5~10", "~2", "20~"}

	str := ns.String()
	for _, part := range expectedParts {
		if !strings.Contains(str, part) {
			t.Errorf("Expected string to contain %q, got: %s", part, str)
		}
	}
}

func TestParseUintegers(t *testing.T) {
	tests := []struct {
		input   string
		wanterr bool
		check   func(Integers) bool
	}{
		{
			input: "1 2 3",
			check: func(us Integers) bool {
				return us.Contains(1) && us.Contains(2) && us.Contains(3)
			},
		},
		{
			input: "1-3",
			check: func(us Integers) bool {
				return us.Contains(1) && us.Contains(2) && us.Contains(3) && !us.Contains(4)
			},
		},
		{
			input: "-3",
			check: func(us Integers) bool {
				return us.Contains(1) && us.Contains(3) && !us.Contains(4)
			},
		},
		{
			input: "5-",
			check: func(us Integers) bool {
				return us.Contains(5) && us.Contains(1000)
			},
		},
		{
			input: "10-10",
			check: func(us Integers) bool {
				return us.Contains(10) && !us.Contains(9)
			},
		},
		{
			input: "20-10",
			check: func(us Integers) bool {
				return us.Contains(10) && us.Contains(20)
			},
		},
		{
			input: " ",
			check: func(us Integers) bool {
				return len(us.ints) == 0 && len(us.ranges) == 0
			},
		},
		{
			input:   "1+a",
			wanterr: true,
		},
		{
			input:   "1 -1+1",
			wanterr: true,
		},
		{
			input:   "1 2+a-",
			wanterr: true,
		},
	}

	for _, tt := range tests {
		us, err := ParseUintegers(tt.input)
		if tt.wanterr {
			if err == nil {
				t.Errorf("ParseUintegers(%q) = %q, but want error", tt.input, us.String())
			}
			continue
		}
		if !tt.check(us) {
			t.Errorf("ParseUintegers(%q) failed", tt.input)
		}
	}
}

func TestUintegersString(t *testing.T) {
	input := "1 3 5-10 -2 20-"
	us, _ := ParseUintegers(input)
	expectedParts := []string{"1", "3", "5-10", "-2", "20-"}

	str := us.String()
	for _, part := range expectedParts {
		if !strings.Contains(str, part) {
			t.Errorf("Expected string to contain %q, got: %s", part, str)
		}
	}
}

// -------------------------------------------------------------------
func TestDecrgContains(t *testing.T) {
	tests := []struct {
		rng    decrg
		n      float64
		expect bool
	}{
		{decrg{nil, float64(10)}, 5, true},
		{decrg{nil, float64(10)}, 15, false},
		{decrg{float64(5), nil}, 6, true},
		{decrg{float64(5), nil}, 3, false},
		{decrg{float64(3), float64(7)}, 5, true},
		{decrg{float64(3), float64(7)}, 8, false},
	}

	for _, tt := range tests {
		got := tt.rng.Contains(tt.n)
		if got != tt.expect {
			t.Errorf("decrg(%v).Contains(%f) = %v, expected %v", tt.rng, tt.n, got, tt.expect)
		}
	}
}

func TestDecrgsContains(t *testing.T) {
	rs := decrgs{
		{float64(1), float64(5)},
		{float64(10), float64(15)},
	}
	if !rs.Contains(3) {
		t.Errorf("Expected Contains(3) = true")
	}
	if rs.Contains(6) {
		t.Errorf("Expected Contains(6) = false")
	}
}

func TestParseDecimals(t *testing.T) {
	tests := []struct {
		input   string
		wanterr bool
		check   func(Decimals) bool
	}{
		{
			input: "-1 2 3",
			check: func(ns Decimals) bool {
				return ns.Contains(-1) && ns.Contains(2) && ns.Contains(3)
			},
		},
		{
			input: "-1~3",
			check: func(ns Decimals) bool {
				return ns.Contains(-1) && ns.Contains(0) && ns.Contains(1) && ns.Contains(2) && ns.Contains(3) && !ns.Contains(4)
			},
		},
		{
			input: "~3",
			check: func(ns Decimals) bool {
				return ns.Contains(1) && ns.Contains(3) && !ns.Contains(4)
			},
		},
		{
			input: "5~",
			check: func(ns Decimals) bool {
				return ns.Contains(5) && ns.Contains(1000)
			},
		},
		{
			input: "10~10",
			check: func(ns Decimals) bool {
				return ns.Contains(10) && !ns.Contains(9)
			},
		},
		{
			input: "20~10",
			check: func(ns Decimals) bool {
				return ns.Contains(10) && ns.Contains(20)
			},
		},
		{
			input: " ",
			check: func(ns Decimals) bool {
				return len(ns.decs) == 0 && len(ns.ranges) == 0
			},
		},
		{
			input:   "1+a",
			wanterr: true,
		},
		{
			input:   "1 ~1+1",
			wanterr: true,
		},
		{
			input:   "1 2+a~",
			wanterr: true,
		},
	}

	for _, tt := range tests {
		ns, err := ParseDecimals(tt.input)
		if tt.wanterr {
			if err == nil {
				t.Errorf("ParseDecimals(%q) = %q, but want error", tt.input, ns.String())
			}
			continue
		}
		if !tt.check(ns) {
			t.Errorf("ParseDecimals(%q) failed", tt.input)
		}
	}
}

func TestDecimalsString(t *testing.T) {
	input := "1 3 5~10 ~2 20~"
	ns, _ := ParseDecimals(input)
	expectedParts := []string{"1", "3", "5~10", "~2", "20~"}

	str := ns.String()
	for _, part := range expectedParts {
		if !strings.Contains(str, part) {
			t.Errorf("Expected string to contain %q, got: %s", part, str)
		}
	}
}

func TestParseUdecimals(t *testing.T) {
	tests := []struct {
		input   string
		wanterr bool
		check   func(Decimals) bool
	}{
		{
			input: "1 2 3",
			check: func(ns Decimals) bool {
				return ns.Contains(1) && ns.Contains(2) && ns.Contains(3)
			},
		},
		{
			input: "1-3",
			check: func(ns Decimals) bool {
				return ns.Contains(1) && ns.Contains(2) && ns.Contains(3) && !ns.Contains(4)
			},
		},
		{
			input: "-3",
			check: func(ns Decimals) bool {
				return ns.Contains(1) && ns.Contains(3) && !ns.Contains(4)
			},
		},
		{
			input: "5-",
			check: func(ns Decimals) bool {
				return ns.Contains(5) && ns.Contains(1000)
			},
		},
		{
			input: "10-10",
			check: func(ns Decimals) bool {
				return ns.Contains(10) && !ns.Contains(9)
			},
		},
		{
			input: "20-10",
			check: func(ns Decimals) bool {
				return ns.Contains(10) && ns.Contains(20)
			},
		},
		{
			input: " ",
			check: func(ns Decimals) bool {
				return len(ns.decs) == 0 && len(ns.ranges) == 0
			},
		},
		{
			input:   "1+a",
			wanterr: true,
		},
		{
			input:   "1 -1+1",
			wanterr: true,
		},
		{
			input:   "1 2+a-",
			wanterr: true,
		},
	}

	for _, tt := range tests {
		ns, err := ParseUdecimals(tt.input)
		if tt.wanterr {
			if err == nil {
				t.Errorf("ParseUdecimals(%q) = %q, but want error", tt.input, ns.String())
			}
			continue
		}
		if !tt.check(ns) {
			t.Errorf("ParseUdecimals(%q) failed", tt.input)
		}
	}
}

func TestUdecimalsString(t *testing.T) {
	input := "1 3 5-10 -2 20-"
	ns, _ := ParseUdecimals(input)
	expectedParts := []string{"1", "3", "5-10", "-2", "20-"}

	str := ns.String()
	for _, part := range expectedParts {
		if !strings.Contains(str, part) {
			t.Errorf("Expected string to contain %q, got: %s", part, str)
		}
	}
}
