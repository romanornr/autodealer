package webserver

import "github.com/thrasher-corp/gocryptotrader/currency"

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
