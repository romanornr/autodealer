package singleton

import (
	"testing"
)

func BenchmarkGetDealerInstanceOnce(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetDealerInstance()
	}
}
