package units

import (
	"math"
)

// FormatMetricBytes formats a MetricBytes value using decimal SI byte units (KB, MB, GB, etc.).
func FormatMetricBytes(m MetricBytes) string {
	return m.String()
}

// FormatMetricBytesSI formats a MetricBytes value using standard SI prefixes (lowercase "kB" for 1000 bytes).
func FormatMetricBytesSI(m MetricBytes) string {
	return ToStringWithPrefixes(int64(m), 1000, siMetricUnits, "B", "B")
}

// FormatBase2Bytes formats a Base2Bytes value using binary byte units (KiB, MiB, GiB, etc.).
func FormatBase2Bytes(b Base2Bytes) string {
	return b.String()
}

// DecimalExponent defines a unit prefix and its corresponding power-of-10 exponent.
type DecimalExponent struct {
	Prefix   string
	Exponent int
}

// MakeDecimalExponentUnitMap creates a unit map for custom decimal exponents (powers of 10).
// Each entry in exponents maps Prefix+suffix to 10^Exponent.
func MakeDecimalExponentUnitMap(suffix, shortSuffix string, exponents []DecimalExponent) map[string]float64 {
	res := map[string]float64{
		shortSuffix: 1,
	}
	for _, exp := range exponents {
		res[exp.Prefix+suffix] = math.Pow(10, float64(exp.Exponent))
	}
	return res
}

// ParseCustomDecimal parses a string into an int64 using custom decimal exponent definitions.
func ParseCustomDecimal(s string, suffix, shortSuffix string, exponents []DecimalExponent) (int64, error) {
	unitMap := MakeDecimalExponentUnitMap(suffix, shortSuffix, exponents)
	return ParseUnit(s, unitMap)
}

// FormatCustomDecimal formats an int64 value with custom exponent prefixes and scale.
func FormatCustomDecimal(n int64, scale int64, prefixes []string, suffix, baseSuffix string) string {
	return ToStringWithPrefixes(n, scale, prefixes, suffix, baseSuffix)
}
