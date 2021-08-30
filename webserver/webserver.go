package webserver

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"html/template"
	"log"
	"net/http"
	"time"
)

var tpl *template.Template

const (
	httpConnTimeout = 10
	port            = 3333
	// notifyRoute is a route used for general notifications.
	notifyRoute = "notify"
	// The basis for content-security-policy. connect-src must be the final
	// directive so that it can be reliably supplemented on startup.
	baseCSP = "default-src 'none'; script-src 'self'; img-src 'self'; style-src 'self'; font-src 'self'; connect-src 'self'"
)

func init() {
	//lvl, _ := logrus.ParseLevel("debug")
	//logrus.SetLevel(lvl)
	tpl = template.Must(template.ParseGlob("site/html/*.html"))
}

// TODO api subrouter
// api/binance/eth/deposit
// api/ftx/eth/deposit
// to get deposit addresses from exchanges easily

// New TODO api withdraw
// api/binance/eth/withdraw/0xethAddress
// the 0xethAddress should be destination
func New() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	chiCors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(chiCors.Handler)

	r.Get("/", HomeHandler)
	r.Get("/deposit", DepositHandler)   // http://127.0.0.1:3333/deposit
	r.Get("/withdraw", WithdrawHandler) // http://127.0.0.1:3333/withdraw
	r.Get("/bank/transfer", bankTransferHandler)

	r.Mount("/api", apiSubrouter())

	logrus.Infof("API route mounted on port %d", port)
	logrus.Infof("creating http Server")

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", port),
		Handler:      r,
		ReadTimeout:  httpConnTimeout * time.Second,
		WriteTimeout: httpConnTimeout * time.Second,
	}

	err := httpServer.ListenAndServe()
	if err != nil {
		logrus.Errorf("failed to start listening: %v", err)
	}
	logrus.Infof("start listening")
}

// A completely separate router for API routes
//
func apiSubrouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		log.Println("brrrrr")
	})

	//r.Get("/accounts", func(w http.ResponseWriter, r *http.Request) {
	//	engine.RESTGetAllEnabledAccountInfo(w, r)
	//})

	r.Route("/deposit/{exchange}/{asset}", func(r chi.Router) {
		r.Use(DepositAddressCtx)
		r.Use(BalanceCtx)
		r.Get("/address", getBalance)
	})

	r.Route("/withdraw/{exchange}/{asset}/{size}/{destinationAddress}", func(r chi.Router) {
		r.Use(WithdrawCtx)
		r.Get("/", getExchangeWithdrawResponse)
	})

	r.Route("/withdraw/history/{exchange}/{asset}", func(r chi.Router) {
		r.Use(withdrawHistoryCtx)
		r.Get("/", getWithdrawHistory)
	})

	r.Route("/ticker/{exchange}/{base}/{quote}", func(r chi.Router) {
		r.Use(TickerCtx)
		r.Get("/", getTicker)
	})

	r.Route("/{exchange}/{base}/priceusd", func(r chi.Router) {
		r.Use(PriceToUSDCtx)
		r.Get("/", getUSDPrice)
	})

	r.Route("/twap/{exchange}/{base}/{quote}/{quantity}", func(r chi.Router) {
		r.Use(TwapCtx)
		r.Get("/", getTwap)
	})

	r.Route("/bank/transfer", func(r chi.Router) {
		r.Use(BankTransferCtx)
		r.Get("/", getBankTransfer)
	})

	return r
}

// HomeHandler handleHome is the handler for the '/' page request. It redirects the
// requester to the markets page.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if err := tpl.ExecuteTemplate(w, "home.html", nil); err != nil {
		logrus.Errorf("error template: %s\n", err)
	}
}

type Asset struct {
	Name       string        `json:"name"`
	Item       currency.Item `json:"item"`
	AssocChain string        `json:"chain"`
	Code       currency.Code `json:"code"`
	Exchange   string        `json:"exchange"`
	Address    string        `json:"address"`
	Balance    string        `json:"balance"`
	Rate       float64       `json:"rate"`
}
