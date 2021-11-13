package webserver

import (
	"github.com/go-chi/chi/v5"
	"github.com/romanornr/autodealer/internal/dealer"
	"github.com/sirupsen/logrus"
	"net/http"
)

// TODO fix HoldingsCtx
func HoldingsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		d := GetDealerInstance()
		_, err := dealer.Holdings(d, chi.URLParam(request, "exchange"))
		if err != nil {
			logrus.Errorf("Error getting holdings: %s\n", err)
		}
	})
}
