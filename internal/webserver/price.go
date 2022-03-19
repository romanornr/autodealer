package webserver

//// fetchPrice returns given the base currency and the target currency
//func fetchPrice(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		base := r.URL.Query().Get("base")
//		quote := r.URL.Query().Get("quote")
//
//		if base == "" || quote == "" {
//			http.Error(w, "Missing base or quote", http.StatusBadRequest)
//			return
//		}
//
//
//
//	}
//
//	return 0, nil
//}
