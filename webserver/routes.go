package webserver

// Routes are API path constants.
const (
	routeGetDepositAddr     = "/deposit/{exchange}/{asset}/{chain}"
	routeWithdraw           = "/withdraw/{exchange}/{asset}/{size}/{destinationAddress}/{chain}"
	routeGetWithdrawHistory = "/withdraw/history/{exchange}/{asset}"
	routeTrade              = "/trade/{exchange}/{pair}"
	routeGetTicker          = "/ticker/{exchange}/{base}/{quote}"
	routePriceToUSD         = "/{exchange}/{base}/priceusd"
	routeTWAP               = "/twap/{exchange}/{base}/{quote}/{quantity}"
	routeBankTransfer       = "/bank/transfer/{currency}"
)
