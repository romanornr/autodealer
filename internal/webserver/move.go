package webserver

import (
	"context"
	"github.com/go-chi/render"
	"github.com/romanornr/autodealer/internal/move"
	"net/http"
)

func MoveTermStructureCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {

		m := move.TermStructure()
		ctx := context.WithValue(request.Context(), "response", m)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}

// getDepositAddress is a function that returns the deposit address for a given exchange and asset.
func getMoveTermStructure(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response, ok := ctx.Value("response").(*depositResponse)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		render.Status(r, http.StatusUnprocessableEntity)
		return
	}
	render.JSON(w, r, response)
}
