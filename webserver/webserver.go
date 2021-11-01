package webserver

import (
	"context"
	"fmt"
	"github.com/romanornr/autodealer/util"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

var tpl *template.Template

const (
	httpConnTimeout = 160
	port            = 3333

	baseCSP = "default-src 'none'; script-src 'self'; img-src 'self'; style-src 'self'; font-src 'self'; connect-src 'self'"
)

// Init sets up our just do for our webserver by ensuring that the ASI Application import has been used correctly.
// The Chi router selects a correct handler and middleware and hooks them together.
func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.Infof(util.Location())
	tpl = template.Must(template.ParseGlob("web/template/*.html"))
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

	// Serve static files from "web" directory
	//workDir, _ := os.Getwd()
	//filesDir := filepath.Join(workDir, "web")
	//http.FileServer(filesDir+"/static")

	r.Get("/", HomeHandler)
	r.Get("/trade", TradeHandler)
	r.Get("/deposit", DepositHandler)   // http://127.0.0.1:3333/deposit
	r.Get("/withdraw", WithdrawHandler) // http://127.0.0.1:3333/withdraw
	r.Get("/bank/transfer", bankTransferHandler)
	r.Get("/s", SearchHandler)

	// func subrouter generates a new router for each sub route.
	r.Mount("/api", apiSubrouter())

	return r
}

// New imports many libraries, effectively constructing the project's "infrastructure."
// These are based on the namespaces' chi, go-chi-middleware, and go-chi-render. Additionally, some little logging was established.
// The remainder of the Routes(), apiSubrouter(), and init() methods configure basic handlers for each resource.
func New() {
	logrus.Infof("API route mounted on port %d", port)
	logrus.Infof("creating http Server")

	go GetDealerInstance()

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", port),
		Handler:      service(),
		ReadTimeout:  httpConnTimeout * time.Second,
		WriteTimeout: httpConnTimeout * (time.Second * 30),
	}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	// Save a reference to our context to be used later
	// serverCtxVar = serverCtx

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logrus.Fatalf("graceful shutdown timeout... forcing exit")
			}
		}()

		// Trigger graceful shutdown
		err := httpServer.Shutdown(shutdownCtx)
		if err != nil {
			logrus.Fatalf("shutdown failed: %v\n", err)
		}
		serverStopCtx()
	}()

	// Run the server
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Fatalf("failed to start listening: %v", err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
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

	r.Route(routeTrade, func(r chi.Router) {
		r.Use(TradeCtx)
		r.Get("/", getTradeResponse)
	})

	r.Route(routeGetDepositAddr, func(r chi.Router) {
		r.Use(DepositAddressCtx)
		r.Get("/", getDepositAddress)
	})

	r.Route(routeWithdraw, func(r chi.Router) {
		r.Use(WithdrawCtx)
		r.Get("/", getExchangeWithdrawResponse)
	})

	r.Route(routeGetWithdrawHistory, func(r chi.Router) {
		r.Use(withdrawHistoryCtx)
		r.Get("/", getWithdrawHistory)
	})

	r.Route(routeBankTransfer, func(r chi.Router) {
		r.Use(BankTransferCtx)
		r.Get("/", getBankTransfer)
	})

	return r
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
