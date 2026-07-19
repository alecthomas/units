package units

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

var (
	siUnits = []string{"", "K", "M", "G", "T", "P", "E"}
)

func ToString(n int64, scale int64, suffix, baseSuffix string) string {
	mn := len(siUnits)
	out := make([]string, mn)

	// Format the unsigned magnitude and prepend a single '-' for negatives.
	// Using the signed value directly emits a '-' per component (n%scale is
	// negative for every unit, e.g. "-9EB-223PB-..."), which is unparseable.
	// The magnitude is held in uint64 so math.MinInt64 is representable (its
	// negation overflows int64).
	neg := n < 0
	mag := uint64(n)
	if neg {
		mag = -mag
	}
	uscale := uint64(scale)

	for i, m := range siUnits {
		if mag%uscale != 0 || i == 0 && mag == 0 {
			s := suffix
			if i == 0 {
				s = baseSuffix
			}
			out[mn-1-i] = fmt.Sprintf("%d%s%s", mag%uscale, m, s)
		}
		mag /= uscale
		if mag == 0 {
			break
		}
	}

	res := strings.Join(out, "")
	if neg {
		res = "-" + res
	}
	return res
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
		if x >= (1<<63-10)/10 {
			// overflow
			return 0, "", errLeadingInt
		}
		x = x*10 + int64(c) - '0'
	}
	return x, s[i:], nil
}

// ParseUnit accumulates with exact rational arithmetic (math/big) rather than
// float64: unit multipliers reach 1000^6 (1e18) and 1024^6 (2^60), both beyond
// the 53-bit float64 mantissa, so a float64 accumulator loses precision near the
// int64 limits and can round MaxInt64 up to MinInt64 on a round-trip.
func ParseUnit(s string, unitMap map[string]int64) (int64, error) {
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	f := new(big.Rat)
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
		g := new(big.Rat) // this element of the sequence

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
		g.SetInt64(x)
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
			scale := int64(1)
			for n := pl - len(s); n > 0; n-- {
				scale *= 10
			}
			g.Add(g, new(big.Rat).SetFrac(big.NewInt(x), big.NewInt(scale)))
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

		g.Mul(g, new(big.Rat).SetInt64(unit))
		f.Add(f, g)
	}

	if neg {
		f.Neg(f)
	}
	// Truncate toward zero (preserving the historical int64(f) behaviour for
	// fractional byte values) with strict overflow detection.
	res := new(big.Int).Quo(f.Num(), f.Denom())
	if !res.IsInt64() {
		return 0, errors.New("units: overflow parsing unit")
	}
	return res.Int64(), nil
}
