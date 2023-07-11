package webserver

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"net/http"
	"os"
	"time"
)

// CorsConfig is our CORS configuration struct
type CorsConfig struct {
	AllowedOrigins   []string // Origins that are allowed to access resources
	AllowedMethods   []string // HTTP Methods that are allowed to be used
	AllowedHeaders   []string // Headers that are allowed in HTTP requests
	ExposedHeaders   []string // Headers that are exposed in HTTP responses
	AllowCredentials bool     // Allow Whether the request can include credentials
	MaxAge           int      // The maximum age (in seconds) of the result of a preflight request
}

// corsMiddleware is our CORS middleware
func corsMiddleware(config *CorsConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return cors.New(cors.Options{
			AllowedOrigins:   config.AllowedOrigins,
			AllowedMethods:   config.AllowedMethods,
			AllowedHeaders:   config.AllowedHeaders,
			ExposedHeaders:   config.ExposedHeaders,
			AllowCredentials: config.AllowCredentials,
			MaxAge:           config.MaxAge,
		}).Handler(next)
	}
}

// setupMiddleware sets up the common middleware used by the router
func (s *Server) setupMiddleware() {
	//logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	//logger = httplog.NewLogger("httplog", httplog.Options{
	//	JSON:            true,
	//	Concise:         true,
	//	LogLevel:        "debug",
	//	TimeFieldFormat: time.TimeOnly,
	//})
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(httplog.RequestLogger(s.logger.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()))
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	// Create the CORS configuration
	corsConfig := &CorsConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           900, // Maximum value not ignored by any of major browsers
	}

	// Set up CORS middleware with the Chi router
	s.router.Use(corsMiddleware(corsConfig))
}
