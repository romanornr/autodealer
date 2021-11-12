package webserver

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/romanornr/autodealer/dealer"
	"github.com/sirupsen/logrus"
	"net/http"
)

func HoldingsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		d := GetDealerInstance()
		holdings, err := dealer.Holdings(d, chi.URLParam(request, "exchange"))
		if err != nil {
			logrus.Errorf("Error getting holdings: %s\n", err)
		}

		ctx := context.WithValue(request.Context(), "response", &holdings)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}
