package webserver

import (
	"context"
	"github.com/go-chi/render"
	"github.com/romanornr/autodealer/internal/move"
	"github.com/sirupsen/logrus"
	"net/http"
)

// getMoveTermStructure returns the move term structure for the given year
func getMoveTermStructure(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response, ok := ctx.Value("response").(*move.TermStructure)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		render.Status(r, http.StatusUnprocessableEntity)
		return
	}
	render.JSON(w, r, response)
}

// MoveTermStructureCtx is a middleware that injects the move term structure into the request context
func MoveTermStructureCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		d := GetDealerInstance()

		m := move.GetTermStructure(d)
		ctx := context.WithValue(request.Context(), "response", &m)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}

// getMoveStats returns the move contracts stats
func getMoveStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response, ok := ctx.Value("response").(*move.List)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		render.Status(r, http.StatusUnprocessableEntity)
		return
	}
	render.JSON(w, r, response)
}

// MoveStatsCtx implements a middleware that injects the move term structure into the request context
func MoveStatsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		d := GetDealerInstance()

		m, err := move.GetStatistics(d)
		if err != nil {
			logrus.Errorf("Error getting move statistics: %v\n", err)
		}
		ctx := context.WithValue(request.Context(), "response", &m)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}
