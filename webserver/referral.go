package webserver

//func getReferral(w http.ResponseWriter, r *http.Request) {
//	ctx := r.Context()
//	response, ok := ctx.Value("response").(*[]ftx.ReferralRebateHistory)
//	if !ok {
//		w.WriteHeader(http.StatusUnprocessableEntity)
//		return
//	}
//
//	w.WriteHeader(http.StatusOK)
//	w.Header().Set("Content-Type", "application/json")
//	render.JSON(w, r, response)
//}
//
//func ReferralCtx(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//
//		var response []ftx.ReferralRebateHistory
//		response = referral.GetFtxReferralLink()
//
//		fmt.Println(response)
//
//		ctx := context.WithValue(r.Context(), "response", &response)
//		next.ServeHTTP(w, r.WithContext(ctx))
//	})
//}
