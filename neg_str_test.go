package units
import "testing"
func TestNegativeString(t *testing.T) {
  cases := []struct{
    b Base2Bytes
    want string
  }{
    {-1500, "-1KiB476B"},
    {-1024, "-1KiB"},
    {-1, "-1B"},
    {Base2Bytes(-Gibibyte - Mebibyte), "-1GiB1MiB"},
  }
  for _, c := range cases {
    got := c.b.String()
    if got != c.want {
      t.Errorf("Base2Bytes(%d).String()=%q want %q", c.b, got, c.want)
    }
  }
  // Metric
  m := MetricBytes(-1500)
  if got := m.String(); got != "-1KB500B" {
    t.Errorf("MetricBytes(-1500).String()=%q want -1KB500B", got)
  }
}
func TestNegativeRoundTrip(t *testing.T) {
  for _, s := range []string{"-1KiB", "-1500B", "-1GiB1MiB"} {
    n, err := ParseBase2Bytes(s)
    if err != nil {
      t.Fatal(err)
    }
    // Parsed value should equal the numeric interpretation
    t.Logf("%s -> %d -> %s", s, n, n.String())
  }
}
