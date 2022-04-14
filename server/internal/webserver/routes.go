package webserver

// Routes are API path constants.
const (
	routeGetDepositAddr     = "/deposit/{exchange}/{asset}/{chain}"
	routeWithdraw           = "/withdraw/{exchange}/{asset}/{size}/{destinationAddress}/{chain}"
	routeGetWithdrawHistory = "/withdraw/history/{exchange}/{asset}"
	routePairs              = "/pairs/{exchange}"
	routeTrade              = "/trade/{exchange}/{pair}/{qty}/{assetType}/{orderType}/{side}"
	routeTWAP               = "/twap/{exchange}/{pair}/{qty}/{assetType}/{orderType}/{side}/{hours}/{minutes}"
	routeGetTicker          = "/ticker/{exchange}/{base}/{quote}"
	routePrice              = "/price/{exchange}/{base}/{quote}/{assetType}"
	routeMoveTermStructure  = "/move"
	routeMoveStats          = "/move/stats"
	routeBankTransfer       = "/bank/transfer/{currency}"
	routeHoldingsExchange   = "/holdings/{exchange}/{asset}"
	routeAssets             = "/assets/{exchange}"
)
