package units

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseUnitInt64Boundary(t *testing.T) {
	for _, input := range []string{"8EiB", "4EiB4EiB", "9EiB", "-9EiB"} {
		t.Run(input, func(t *testing.T) {
			_, err := ParseUnit(input, bytesUnitMap)
			assert.EqualError(t, err, "units: overflow parsing unit")
		})
	}
	for _, tc := range []struct {
		input string
		want  int64
	}{
		{"7EiB", 7 << 60},
		{"-7EiB", -7 << 60},
		{"-8EiB", -1 << 63},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseUnit(tc.input, bytesUnitMap)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
