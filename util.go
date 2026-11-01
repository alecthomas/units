package units

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

var (
	siUnits = []string{"", "K", "M", "G", "T", "P", "E"}
)

func ToString(n int64, scale int64, suffix, baseSuffix string) string {
	mn := len(siUnits)
	out := make([]string, mn)
	for i, m := range siUnits {
		if n%scale != 0 || i == 0 && n == 0 {
			s := suffix
			if i == 0 {
				s = baseSuffix
			}
			out[mn-1-i] = fmt.Sprintf("%d%s%s", n%scale, m, s)
		}
		n /= scale
		if n == 0 {
			break
		}
	}
	return strings.Join(out, "")
}

// Below code ripped straight from http://golang.org/src/pkg/time/format.go?s=33392:33438#L1123
var errLeadingInt = errors.New("units: bad [0-9]*") // never printed

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x int64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
	// Check for overflow before x = x*10 + digit, taking the current
	// digit into account, so that MaxInt64 itself is still accepted.
	if x > (1<<63-1-int64(c)+'0')/10 {
			// overflow
			return 0, "", errLeadingInt
		}
		x = x*10 + int64(c) - '0'
	}
	return x, s[i:], nil
}

func ParseUnit(s string, unitMap map[string]float64) (int64, error) {
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	f := float64(0) // accumulates fractional contributions
	n := int64(0)   // accumulates integral contributions exactly
	neg := false

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}
	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, nil
	}
	if s == "" {
		return 0, errors.New("units: invalid " + orig)
	}
	for s != "" {
		g := float64(0) // this element of the sequence

		var x int64
		var err error

		// The next character must be [0-9.]
		if !(s[0] == '.' || ('0' <= s[0] && s[0] <= '9')) {
			return 0, errors.New("units: invalid " + orig)
		}
		// Consume [0-9]*
		pl := len(s)
		x, s, err = leadingInt(s)
		if err != nil {
			return 0, errors.New("units: invalid " + orig)
		}
		xint := x
		pre := pl != len(s) // whether we consumed anything before a period

		// Consume (\.[0-9]*)?
		post := false
		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			x, s, err = leadingInt(s)
			if err != nil {
				return 0, errors.New("units: invalid " + orig)
			}
			scale := 1.0
			for n := pl - len(s); n > 0; n-- {
				scale *= 10
			}
			g = float64(x) / scale
			post = pl != len(s)
		}
		if !pre && !post {
			// no digits (e.g. ".s" or "-.s")
			return 0, errors.New("units: invalid " + orig)
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == '.' || ('0' <= c && c <= '9') {
				break
			}
		}
		u := s[:i]
		s = s[i:]
		unit, ok := unitMap[u]
		if !ok {
			return 0, errors.New("units: unknown unit " + u + " in " + orig)
		}

		// Accumulate integral contributions with exact integer arithmetic.
		// The unit multipliers reach 1024^6 and 1000^6 (~2^60), so
		// accumulating in float64 silently loses precision above 2^53.
		// Fractional parts, and non-integral unit multipliers from custom
		// unit maps, are still accumulated in float64.
		if m, integral := integralUnit(unit); integral {
			var ok bool
			if n, ok = mulAddInt64(n, xint, m); !ok {
				return 0, errors.New("units: overflow parsing unit")
			}
			f += g * unit
		} else {
			f += (float64(xint) + g) * unit
		}
	}

	if neg {
		n = -n
		f = -f
	}
	// The fractional accumulator must be within int64 range before it is
	// truncated towards zero and added to the integral part.
	if f >= float64(1<<63) || f < float64(-1<<63) {
		return 0, errors.New("units: overflow parsing unit")
	}
	r := n + int64(f)
	if i := int64(f); (i > 0 && r < n) || (i < 0 && r > n) {
		return 0, errors.New("units: overflow parsing unit")
	}
	return r, nil
}

// integralUnit returns unit as an int64 if it is integral, non-negative and
// within int64 range.
func integralUnit(unit float64) (int64, bool) {
	if unit < 0 || unit >= float64(1<<63) || unit != math.Trunc(unit) {
		return 0, false
	}
	return int64(unit), true
}

// mulAddInt64 returns a + x*m, reporting whether the operation overflowed.
// x, produced by leadingInt, and m, produced by integralUnit, are both
// non-negative.
func mulAddInt64(a, x, m int64) (int64, bool) {
	if x == 0 || m == 0 {
		return a, true
	}
	if x > (1<<63-1)/m {
		return 0, false
	}
	p := x * m
	r := a + p
	if r < a { // p > 0
		return 0, false
	}
	return r, true
}
