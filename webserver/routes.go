package webserver

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

// Routes are API path constants.
const (
	routeAvailableTransferChains = "/transfer/chains/{exchange}/{asset}"
	routeGetDepositAddr          = "/deposit/{exchange}/{asset}/{chain}"
	routeWithdraw                = "/withdraw/{exchange}/{asset}/{size}/{destinationAddress}/{chain}"
	routeGetWithdrawHistory      = "/withdraw/history/{exchange}/{asset}"
	routePairs                   = "/pairs/{exchange}"
	routeTrade                   = "/trade/{exchange}/{pair}/{qty}/{assetType}/{orderType}/{side}"
	routeTWAP                    = "/twap/{exchange}/{pair}/{qty}/{assetType}/{orderType}/{side}/{hours}/{minutes}"
	routeGetTicker               = "/ticker/{exchange}/{base}/{quote}"
	routePrice                   = "/price/{exchange}/{base}/{quote}/{assetType}"
	routeMoveTermStructure       = "/move"
	routeMoveStats               = "/move/stats"
	routeBankTransfer            = "/bank/transfer/{currency}"
	routeHoldingsExchange        = "/holdings/{exchange}/{asset}"
	routeAssets                  = "/assets/{exchange}"
	routeReferral                = "/referral"
)

// SetupRoutes configures the HTTP routes for the server. It takes a Handler object
// which contains methods for handling different paths. It returns a configured chi.Mux router object.
// The function maps URL paths to handler methods. For example, it maps the path "/"
// to the handler's "handleTemplate" method with "home.html" as a parameter.
// It also mounts a subrouter for the "/api" path. This allows for a separate set of routes to be defined for API requests.
// This method should be called as part of the server setup process before starting the server.
func (s *Server) SetupRoutes(handler *Handler) *chi.Mux {
	s.router.Get("/", handler.handleTemplate("home.html")) // handler.handleTemplate("index"))
	s.router.Get("/test", handler.handleTemplate("bank.html"))
	s.router.Get("/trade", handler.handleTemplate("trade.html"))       //handler.TradeHandler)
	s.router.Get("/deposit", handler.handleTemplate("deposit.html"))   // http://127.0.0.1:3333/deposit
	s.router.Get("/withdraw", handler.handleTemplate("withdraw.html")) // http://127.0.0.1:3333/withdraw
	s.router.Get("/bank/transfer", handler.handleTemplate("bank.html"))
	s.router.Get("/s", handler.handleTemplate("search.html"))
	//r.Get("/move", MoveHandler) // http://127.0.0.1:3333/move

	// func subrouter generates a new router for each sub route.
	s.router.Mount("/api", apiSubrouter())

	return s.router
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
