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
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
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

type Handler struct {
	tpl    *template.Template
	logger *zerolog.Logger
}

func NewHandler() *Handler {
	return &Handler{
		tpl:    template.Must(template.ParseGlob("webserver/templates/*.html")),
		logger: initLogger(),
	}
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

func (s *Server) SetupService() http.Handler {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	logger = httplog.NewLogger("httplog", httplog.Options{
		JSON:            true,
		Concise:         true,
		LogLevel:        "debug",
		TimeFieldFormat: time.TimeOnly,
	})

	// Set up our middleware with the Chi router
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(httplog.RequestLogger(logger))
	s.router.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further processing should be stopped.
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

func (s *Server) SetupRoutes() *chi.Mux {
	handler := NewHandler() // call NewHandler() to get a new handler instance

	s.router.Get("/", handler.HomeHandler)
	s.router.Get("/trade", handler.TradeHandler)
	s.router.Get("/deposit", handler.DepositHandler)   // http://127.0.0.1:3333/deposit
	s.router.Get("/withdraw", handler.WithdrawHandler) // http://127.0.0.1:3333/withdraw
	s.router.Get("/bank/transfer", handler.bankTransferHandler)
	s.router.Get("/s", handler.SearchHandler)
	//r.Get("/move", MoveHandler) // http://127.0.0.1:3333/move

	// func subrouter generates a new router for each sub route.
	s.router.Mount("/api", apiSubrouter())

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

	s.server.Handler = s.SetupService() // set up the service
	s.router = s.SetupRoutes()

	// initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error().Msgf("error starting server: %s", err)
		}
		s.logger.Info().Msg("server stopped listening")
	}()
	return s, nil
}

func New() {

	// Create context that listns for the interrupt signal.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

	// Listen for the interrupt signal
	<-ctx.Done()

	// Restore default behavior on interrupt signal and notify user of shutdown.
	stop()
	s.logger.Info().Msgf("shutting down server grafecully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = s.server.Shutdown(ctx); err != nil {
		s.logger.Error().Msgf("error shutting down http server: %s\n", err)
	}

	s.logger.Info().Msg("server exiting")
}

// apiSubrouter function will create an api route tree for each exchange, which will then be mounted into the application routing tree using the apiSubroutines.Mount method.
// It will then apply the WithdrawCtx function to any API requests that include the /withdraw, /deposit, or /twap routes. These three features are included in sendRequestSpecific.
func apiSubrouter() http.Handler {
	r := chi.NewRouter()

	r.Route(routePairs, func(r chi.Router) {
		r.Use(FetchPairsCtx)
		r.Get("/", getPairsResponse)
	})

	r.Route(routePrice, func(r chi.Router) {
		r.Use(PriceCtx)
		r.Get("/", getPrice)
	})

	r.Route(routeTrade, func(r chi.Router) {
		r.Use(TradeCtx)
		r.Get("/", getTradeResponse)
	})

	r.Route(routeHoldingsExchange, func(r chi.Router) {
		r.Use(HoldingsExchangeCtx)
		r.Get("/", getHoldingsExchangeResponse)
	})

	r.Route(routeAvailableTransferChains, func(r chi.Router) {
		r.Use(AvailableTransferChainsCtx)
		r.Get("/", getAvailableTransferChainsResponse)
	})

	r.Route(routeGetDepositAddr, func(r chi.Router) {
		r.Use(DepositAddressCtx)
		r.Get("/", getDepositAddress)
	})

	r.Route(routeWithdraw, func(r chi.Router) {
		r.Use(WithdrawCtx)
		r.Get("/", getExchangeWithdrawResponse)
	})

	r.Route(routeBankTransfer, func(r chi.Router) {
		r.Use(BankTransferCtx)
		r.Get("/", getBankTransfer)
	})

	r.Route(routeAssets, func(r chi.Router) {
		r.Use(AssetListCtx)
		r.Get("/", getAssetList)
	})

	r.Route(routeTWAP, func(r chi.Router) {
		r.Use(TWAPCtx)
		r.Get("/", getTwapResponse)
	})
	return r
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

func (h *Handler) HomeHandler(w http.ResponseWriter, _ *http.Request) {
	if err := h.tpl.ExecuteTemplate(w, "home.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error().Msgf("error template: %s\n", err)
		return
	}
}

func (h *Handler) SearchHandler(w http.ResponseWriter, _ *http.Request) {
	if err := h.tpl.ExecuteTemplate(w, "search.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error().Msgf("error template: %s\n", err)
		return
	}
}
