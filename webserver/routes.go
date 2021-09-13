package webserver

// Routes are API path constants.
const (
	routeGetDepositAddr     = "/deposit/{exchange}/{asset}"
	routeWithdraw           = "/withdraw/{exchange}/{asset}/{size}/{destinationAddress}"
	routeGetWithdrawHistory = "/withdraw/history/{exchange}/{asset}"
	routeGetTicker          = "/ticker/{exchange}/{base}/{quote}"
	routePriceToUSD         = "/{exchange}/{base}/priceusd"
	routeTWAP               = "/twap/{exchange}/{base}/{quote}/{quantity}"
	routeBankTransfer       = "/bank/transfer"
)
