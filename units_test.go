package units

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatMetricBytes(t *testing.T) {
	assert.Equal(t, "0B", FormatMetricBytes(MetricBytes(0)))
	assert.Equal(t, "1KB", FormatMetricBytes(KB))
	assert.Equal(t, "1KB1B", FormatMetricBytes(MetricBytes(1001)))
	assert.Equal(t, "1MB1KB25B", FormatMetricBytes(MetricBytes(1001025)))
	assert.Equal(t, "1EB", FormatMetricBytes(EB))
}

func TestFormatMetricBytesSI(t *testing.T) {
	assert.Equal(t, "0B", FormatMetricBytesSI(MetricBytes(0)))
	assert.Equal(t, "1kB", FormatMetricBytesSI(MetricBytes(1000)))
	assert.Equal(t, "1kB1B", FormatMetricBytesSI(MetricBytes(1001)))
	assert.Equal(t, "1MB1kB25B", FormatMetricBytesSI(MetricBytes(1001025)))
	assert.Equal(t, "1MB500kB", FormatMetricBytesSI(MetricBytes(1500000)))

	// Verify method equality
	assert.Equal(t, "1kB", MetricBytes(1000).FormatSI())
	assert.Equal(t, "1MB1kB25B", MetricBytes(1001025).FormatSI())
}

func TestFormatBase2Bytes(t *testing.T) {
	assert.Equal(t, "0B", FormatBase2Bytes(Base2Bytes(0)))
	assert.Equal(t, "1KiB", FormatBase2Bytes(KiB))
	assert.Equal(t, "1KiB1B", FormatBase2Bytes(Base2Bytes(1025)))
	assert.Equal(t, "1MiB1KiB25B", FormatBase2Bytes(Base2Bytes(1049625)))
	assert.Equal(t, "1EiB", FormatBase2Bytes(EiB))
}

func TestMakeDecimalExponentUnitMap(t *testing.T) {
	exponents := []DecimalExponent{
		{Prefix: "k", Exponent: 3},
		{Prefix: "M", Exponent: 6},
		{Prefix: "G", Exponent: 9},
	}
	m := MakeDecimalExponentUnitMap("B", "B", exponents)
	assert.Equal(t, float64(1), m["B"])
	assert.Equal(t, float64(1000), m["kB"])
	assert.Equal(t, float64(1000000), m["MB"])
	assert.Equal(t, float64(1000000000), m["GB"])
}

func TestParseCustomDecimal(t *testing.T) {
	exponents := []DecimalExponent{
		{Prefix: "m", Exponent: -3},
		{Prefix: "c", Exponent: -2},
		{Prefix: "d", Exponent: -1},
		{Prefix: "da", Exponent: 1},
		{Prefix: "h", Exponent: 2},
		{Prefix: "k", Exponent: 3},
		{Prefix: "M", Exponent: 6},
	}

	val, err := ParseCustomDecimal("0m", "m", "m", exponents)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), val)

	val, err = ParseCustomDecimal("5km", "m", "m", exponents)
	assert.NoError(t, err)
	assert.Equal(t, int64(5000), val)

	val, err = ParseCustomDecimal("1Mm2km", "m", "m", exponents)
	assert.NoError(t, err)
	assert.Equal(t, int64(1002000), val)

	val, err = ParseCustomDecimal("-2km", "m", "m", exponents)
	assert.NoError(t, err)
	assert.Equal(t, int64(-5000+3000), val) // -2000

	val, err = ParseCustomDecimal("1000mm", "m", "m", exponents)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	// Invalid input
	_, err = ParseCustomDecimal("invalid", "m", "m", exponents)
	assert.Error(t, err)

	// Unknown unit
	_, err = ParseCustomDecimal("5xm", "m", "m", exponents)
	assert.Error(t, err)
}

func TestFormatCustomDecimal(t *testing.T) {
	prefixes := []string{"", "da", "h", "k"}
	assert.Equal(t, "0m", FormatCustomDecimal(0, 10, prefixes, "m", "m"))
	assert.Equal(t, "1km", FormatCustomDecimal(1000, 10, prefixes, "m", "m"))
	assert.Equal(t, "1km2hm3dam4m", FormatCustomDecimal(1234, 10, prefixes, "m", "m"))

	// Edge cases
	assert.Equal(t, "42m", FormatCustomDecimal(42, 0, prefixes, "m", "m"))
	assert.Equal(t, "42m", FormatCustomDecimal(42, -10, prefixes, "m", "m"))
	assert.Equal(t, "42m", FormatCustomDecimal(42, 10, []string{}, "m", "m"))
}
