package needle

import (
	"testing"
)

func TestGetNeedle(t *testing.T) {
	roids := GetRoids()
	secondRoids := GetRoids()
	if roids != secondRoids {
		t.Error("Both roids should be the same instance.")
	}
}
