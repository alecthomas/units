package units

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetricBytesMarshalCycle covers two defects around the int64 limits:
//
//  1. ParseUnit accumulated into a float64. Unit multipliers reach 1e18 and
//     2^60, beyond the 53-bit float64 mantissa, so MaxInt64 rounded up on parse
//     and int64() wrapped it to MinInt64.
//  2. ToString formatted each component from a signed value, emitting a '-' per
//     unit for negatives (e.g. "-9EB-223PB-..."), a string it could not itself
//     parse.
func TestMetricBytesMarshalCycle(t *testing.T) {
	values := []int64{
		0,
		1,
		-1,
		1000,
		-1000,
		math.MaxInt64,
		math.MinInt64,
	}

	for _, v := range values {
		s := MetricBytes(v).String()

		back, err := ParseStrictBytes(s)
		require.NoErrorf(t, err, "failed to parse %q (from %d)", s, v)
		assert.Equalf(t, v, back, "round-trip mismatch via %q", s)
	}
}

// TestToStringSingleSign guards defect (2): a negative value carries exactly one
// leading '-', not one per unit component.
func TestToStringSingleSign(t *testing.T) {
	assert.Equal(t, "-9EB223PB372TB36GB854MB775KB808B", MetricBytes(math.MinInt64).String())
	assert.Equal(t, "9EB223PB372TB36GB854MB775KB807B", MetricBytes(math.MaxInt64).String())
	assert.Equal(t, "-1KB1B", MetricBytes(-1001).String())
	assert.Equal(t, "-1KiB1B", Base2Bytes(-1025).String())
}

// TestParseNoFloatOverflow guards defect (1): the exact MaxInt64 string parses
// back to MaxInt64 instead of overflowing to MinInt64.
func TestParseNoFloatOverflow(t *testing.T) {
	n, err := ParseStrictBytes("9EB223PB372TB36GB854MB775KB807B")
	require.NoError(t, err)
	assert.Equal(t, int64(math.MaxInt64), n)
}
