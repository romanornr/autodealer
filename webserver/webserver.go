package webserver

import "C"
import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/romanornr/autodealer/util"
	"html/template"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/romanornr/autodealer/algo/twap"
	"github.com/romanornr/autodealer/config"
	"github.com/romanornr/autodealer/singleton"
	"github.com/spf13/viper"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

var tpl *template.Template

const (
	redisAddr = "127.0.0.1:6379"
	baseCSP   = "default-src 'none'; script-src 'self'; img-src 'self'; style-src 'self'; font-src 'self'; connect-src 'self'"
)

// Init sets up our just do for our webserver by ensuring that the ASI Application import has been used correctly.
// The Chi router selects a correct handler and middleware and hooks them together.
func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.Infof(util.Location())

	// time format YYYY-MM-DD HH:MM:SS
	logrus.Infof("%s %s", time.Now().Format("2006-01-02 15:04:05"), util.Location()+": Init")

	tpl = template.Must(template.ParseGlob("webserver/templates/*.html"))
}

func service() http.Handler {

	r := chi.NewRouter()

	// The middleware stacks. Logger, per RequestId and re-hopping initialized variables.
	// The RequestId middleware handles uuid generation for each request and setting it to Mux context.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	chiCors := corsConfig()

	r.Use(chiCors.Handler)

	// set 404 handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if err := tpl.ExecuteTemplate(w, "404.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logrus.Errorf("error template: %s\n", err)
		}
	})

	r.Get("/", HomeHandler)
	r.Get("/trade", TradeHandler)
	r.Get("/deposit", DepositHandler)   // http://127.0.0.1:3333/deposit
	r.Get("/withdraw", WithdrawHandler) // http://127.0.0.1:3333/withdraw
	r.Get("/bank/transfer", bankTransferHandler)
	r.Get("/s", SearchHandler)
	//r.Get("/move", MoveHandler) // http://127.0.0.1:3333/move

	// func subrouter generates a new router for each sub route.
	r.Mount("/api", apiSubrouter())

	return r
}

// New imports many libraries, effectively constructing the project's "infrastructure."
// These are based on the namespaces' chi, go-chi-middleware, and go-chi-render. Additionally, some little logging was established.
// The remainder of the Routes(), apiSubrouter(), and init() methods configure basic handlers for each resource.
// TODO We can improve the project's performance by using the chi.Mux.StrictSlash(true) option.
func New() {
	// Create context that listns for the interrupt signal.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// load config
	config.LoadAppConfig()
	logrus.Infof("API route mounted on port %s\n", viper.GetString("SERVER_PORT"))
	logrus.Infof("creating http Server")

	//go singleton.singleton.GetDealerInstance()
	go singleton.GetDealer()
	go asyncWebWorker()

	httpServer := &http.Server{
		// viper config .env get server address
		Addr:           viper.GetViper().GetString("SERVER_ADDR") + ":" + viper.GetViper().GetString("SERVER_PORT"),
		Handler:        service(),
		ReadTimeout:    viper.GetViper().GetDuration("SERVER_READ_TIMEOUT"),
		WriteTimeout:   viper.GetViper().GetDuration("SERVER_WRITE_TIMEOUT"),
		MaxHeaderBytes: 1 << 20,
	}

	// initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("error starting http server: %s\n", err)
		}
		logrus.Printf("server stopped serving new connections")
	}()

	// Listen for the interrupt signal
	<-ctx.Done()

	// Restore default behavior on interrupt signal and notify user of shutdown.
	stop()
	logrus.Infof("shutting downserver gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logrus.Errorf("error shutting down http server: %s\n", err)
	}

	logrus.Infof("server exiting")
}

// The `corsConfig` function returns a new CORS configuration. It is used to configure CORS for our application. The CORS configuration is used by the `cors.New` middleware.
func corsConfig() *cors.Cors {
	return cors.New(cors.Options{
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
	})
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

// HomeHandler handleHome is the handler for the '/' page request. It redirects the
// requester to the markets page.
func HomeHandler(w http.ResponseWriter, _ *http.Request) {
	if err := tpl.ExecuteTemplate(w, "home.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.Errorf("error template: %s\n", err)
		return
	}
}

func SearchHandler(w http.ResponseWriter, _ *http.Request) {
	if err := tpl.ExecuteTemplate(w, "search.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.Errorf("error template: %s\n", err)
		return
	}
}
