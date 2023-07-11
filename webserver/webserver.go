package webserver

import "C"
import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog"
	"github.com/hibiken/asynq"
	"github.com/romanornr/autodealer/algo/twap"
	"github.com/romanornr/autodealer/config"
	"github.com/romanornr/autodealer/singleton"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	redisAddr = "127.0.0.1:6379"
	baseCSP   = "default-src 'none'; script-src 'self'; img-src 'self'; style-src 'self'; font-src 'self'; connect-src 'self'"
)

type Server struct {
	server *http.Server
	logger *zerolog.Logger
	router *chi.Mux
}

func initTpl() *template.Template {
	tpl, err := template.ParseGlob("webserver/templates/*html") ///template.Must(template.ParseGlob("webserver/templates/*.html"))
	if err != nil {
		panic(err)
	}
	return tpl
	//return template.Must(template.ParseGlob("webserver/templates/*.html"))
}

func initLogger() *zerolog.Logger {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return &logger
}

// Init sets up our just do for our webserver by ensuring that the ASI Application import has been used correctly.
// The Chi router selects a correct handler and middleware and hooks them together.
func init() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	config.LoadAppConfig()
}

func (s *Server) SetupService() http.Handler {
	// Set up our middleware with the Chi router
	s.setupMiddleware() // Call to setupMiddleware function

	handler := NewHandler()

	// set 404 handler
	s.router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		oplog := httplog.LogEntry(r.Context())
		w.WriteHeader(http.StatusNotFound)
		oplog.Warn().Msgf("path not found: %q", r.URL.Path)
		if err := handler.tpl.ExecuteTemplate(w, "404.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			s.logger.Error().Msgf("error template: %s\n", err)
		}
	})

	return s.router
}

// NewServer creates a new HTTP server and returns a pointer to it.
func NewServer() (*Server, error) {
	s := &Server{
		server: &http.Server{
			Addr:           viper.GetViper().GetString("SERVER_ADDR") + ":" + viper.GetViper().GetString("SERVER_PORT"),
			ReadTimeout:    viper.GetViper().GetDuration("SERVER_READ_TIMEOUT"),
			WriteTimeout:   viper.GetViper().GetDuration("SERVER_WRITE_TIMEOUT"),
			MaxHeaderBytes: 1 << 20,
		},
		logger: initLogger(),
		router: chi.NewRouter(), // initialize the chi router
	}

	handler := NewHandler()
	s.server.Handler = s.SetupService() // set up the service
	s.router = s.SetupRoutes(handler)

	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	startTime = time.Now()
	go func() {
		if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error().Msgf("error starting server: %s", err)
		}
		s.logger.Info().Msg("server stopped listening")
	}()

	// Listen for the interrupt signal
	<-ctx.Done()
	s.logger.Info().Msg("shutting down server")

	// The context is used to inform the server it has 5 seconds to finish
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error().Msgf("error shutting down http server: %s", err)
		return err
	}

	s.logger.Info().Msg("server exiting")
	return nil
}

func New(ctx context.Context) {
	s, err := NewServer()
	if err != nil {
		s.logger.Fatal().Msgf("error creating server: %s", err)
	}

	s.logger.Info().Msgf("config loaded: %s", viper.ConfigFileUsed())
	s.logger.Info().Msgf("API route mounted on port %s\n", viper.GetString("SERVER_PORT"))
	s.logger.Info().Msg("creating http server")

	//go singleton.singleton.GetDealerInstance()
	go singleton.GetDealer()
	go asyncWebWorker()

	// Run the server and handle shutdown
	err = s.Start(ctx)
	if err != nil {
		s.logger.Error().Msgf("error running server: %s", err)
	}
}

func asyncWebWorker() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// See the godoc for other configuration options
		},
	)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(twap.TypeTwap, twap.HandleTwapTask) ///  TODO find url 127.0.0.1:3333/twap ??
	mux.HandleFunc(twap.TypeOrder, twap.HandleOrderTask)
	// ...register other handlers...

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
