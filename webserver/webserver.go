package webserver

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
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

	// notifyRoute is a route used for general notifications.
	notifyRoute = "notify"

	// The basis for content-security-policy. connect-src must be the final
	// directive so that it can be reliably supplemented on startup.
	baseCSP = "default-src 'none'; script-src 'self'; img-src 'self'; style-src 'self'; font-src 'self'; connect-src 'self'"
)

// Init sets up our just do for our webserver by ensuring that the ASI Application import has been used correctly.
// The Chi router selects a correct handler and middleware and hooks them together.
func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})
	tpl = template.Must(template.ParseGlob("web/template/*.html"))
}

// New imports many libraries, effectively constructing the project's "infrastructure."
// These are based on the namespaces' chi, go-chi-middleware, and go-chi-render. Additionally, some little logging was established.
// The remainder of the Routes(), apiSubrouter(), and init() methods configure basic handlers for each resource.
func New() {
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

	chiCors := cors.New(cors.Options{
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

	r.Use(chiCors.Handler)

	r.Get("/", HomeHandler)
	r.Get("/deposit", DepositHandler)   // http://127.0.0.1:3333/deposit
	r.Get("/withdraw", WithdrawHandler) // http://127.0.0.1:3333/withdraw
	r.Get("/bank/transfer", bankTransferHandler)

	// func subrouter generates a new router for each sub route.
	r.Mount("/api", apiSubrouter())

	logrus.Infof("API route mounted on port %d", port)
	logrus.Infof("creating http Server")

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", port),
		Handler:      r,
		ReadTimeout:  httpConnTimeout * time.Second,
		WriteTimeout: httpConnTimeout * (time.Second * 30),
	}

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Fatalf("failed to start listening: %v", err)
	}
}

// apiSubrouter function will create an api route tree for each exchange, which will then be mounted into the application routing tree using the apiSubroutines.Mount method.
// It will then apply the WithdrawCtx function to any API requests that include the /withdraw, /deposit, or /twap routes. These three features are included in sendRequestSpecific.
func apiSubrouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		log.Println("brrrrr")
	})

	//r.Get("/accounts", func(w http.ResponseWriter, r *http.Request) {
	//	engine.RESTGetAllEnabledAccountInfo(w, r)
	//})

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

	//r.Route(routeGetTicker, func(r chi.Router) {
	//	r.Use(TickerCtx)
	//	r.Get("/", getTicker)
	//})
	//
	//r.Route(routePriceToUSD, func(r chi.Router) {
	//	r.Use(PriceToUSDCtx)
	//	r.Get("/", getUSDPrice)
	//})
	//
	//r.Route(routeTWAP, func(r chi.Router) {
	//	r.Use(TwapCtx)
	//	r.Get("/", getTwap)
	//})

	r.Route(routeBankTransfer, func(r chi.Router) {
		r.Use(BankTransferCtx)
		r.Get("/", getBankTransfer)
	})

	return r
}

// AcceptHeaderCtx chi/middleware-package offers a wrapper for existing functions and provides Controller patterns based around to Accept header of the request, parsing and returning it to our Router
//func AcceptHeaderCtx(next http.Handler) {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		accept := r.Header.Get("Accept")
//		switch accept {
//		case "application/json":
//			next.ServeHTTP(w, r)
//		default:
//			w.Header().Set("Content-Type", "application/json")
//		}
//	}
//}


// HomeHandler handleHome is the handler for the '/' page request. It redirects the
// requester to the markets page.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if err := tpl.ExecuteTemplate(w, "home.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.Errorf("error template: %s\n", err)
		return
	}
}
