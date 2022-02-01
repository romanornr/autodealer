package webserver

import "testing"

//// Benchmark GetDealerInstance
//func BenchmarkGetDealerInstance(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		GetDealerInstance()
//	}
//}

func BenchmarkGetDealerInstanceOnce(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetDealerInstanceOnce()
	}
}
