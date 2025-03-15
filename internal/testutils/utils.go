package testutils

import (
	"testing"

	"github.com/govalues/decimal"
)

func NewDecimal(t testing.TB, value int64, scale int) decimal.Decimal {
	t.Helper()
	d, err := decimal.New(value, scale)
	if err != nil {
		t.Errorf("NewDecimal(%d, %d) failed: %v", value, scale, err)
	}
	return d
}
