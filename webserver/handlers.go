package webserver

import (
	"github.com/rs/zerolog"
	"html/template"
	"net/http"
)

// Handler is a struct that encapsulates the resources needed to handle HTTP requests.
// It contains a `tpl` field for HTML templates that can be used to render web pages,
// and a `logger` field from the zerolog package for logging information about the requests and responses.
// Handler methods can use these resources to handle different HTTP requests. For example, the `handleTemplate` method uses the `tpl` field to render HTML pages.
// A new Handler can be created using the `NewHandler` function, which initializes the templates and logger.
type Handler struct {
	tpl    *template.Template
	logger *zerolog.Logger
}

// NewHandler creates and returns a new Handler. It initializes templates and logger for the handler.
func NewHandler() *Handler {
	return &Handler{
		tpl:    initTpl(), //template.Must(template.ParseGlob("webserver/templates/*.html")),
		logger: initLogger(),
	}
}

// handleTemplate will render the template with the name passed as parameter.
// It will return a http.HandlerFunc which can be used to handle the request.
// handleTemplate function is a wrapper for the template handler functions
func (h *Handler) handleTemplate(templateName string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// common handler code here...
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := h.tpl.ExecuteTemplate(w, templateName, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			h.logger.Error().Msgf("error template: %s\n", err)
			return
		}
	}
}
