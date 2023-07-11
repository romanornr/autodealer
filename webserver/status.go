package webserver

import (
	"github.com/go-chi/render"
	"net/http"
	"time"
)

var startTime time.Time

type Status struct {
	Status string `json:"status"`
	Uptime string `json:"uptime"`
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)
	status := Status{
		Status: "running",
		Uptime: uptime.String(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, status)
}
