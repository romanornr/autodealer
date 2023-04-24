package webserver

import (
	"github.com/sirupsen/logrus"
	"strings"
)

type ChainSelector interface {
	SelectChain(chainReq string, availableTransferChains []string) string
}

type BinanceChainSelector struct{}
type FTXChainSelector struct{}
type HuobiChainSelector struct{}
type KrakenCHainSelector struct{}
type BTSEChainSelector struct{}
type BittrexChainSelector struct{}

func (c *BinanceChainSelector) SelectChain(chainReq string, availableTransferChains []string) string {
	switch chainReq {
	case "erc20":
		return "ETH"
	case "trx":
		return "TRX"
	case "sol":
		return "SOL"
	default:
		return ""
	}
}

func (c *FTXChainSelector) SelectChain(chainReq string, availableTransferChains []string) string {
	switch chainReq {
	case "erc20":
		return "erc20"
	case "trx":
		return "trx"
	case "sol":
		return "sol"
	default:
		return ""
	}
}

func (c *BittrexChainSelector) SelectChain(chainReq string, availableTransferChains []string) string {
	switch chainReq {
	case "erc20":
		return "erc20"
	case "trx":
		return "trx"
	case "sol":
		return "sol"
	default:
		return ""
	}
}

func GetChainSelector(exchangeName string) ChainSelector {
	switch exchangeName {
	case "Binance":
		return &BinanceChainSelector{}
	case "FTX":
		return &FTXChainSelector{}
	//case "Huobi":
	//    return &HuobiChainSelector{}
	//case "Kraken":
	//    return &KrakenCHainSelector{}
	//case "BTSE":
	//    return &BTSEChainSelector{}
	case "Bittrex":
		return &BittrexChainSelector{}
	default:
		return nil
	}
}

func chainSelection(exchangeName string, chainReq string, availableTransferChains []string) string {
	selector := GetChainSelector(strings.ToLower(exchangeName))
	if selector == nil {
		logrus.Warnf("Chain Selection error: Unsupported exchange: %s", exchangeName)
	}

	if len(availableTransferChains) == 0 {
		logrus.Warnf("Chain Selection error: No multiple available transfer chains for exchange: %s")
		//return selector.SelectChain(chainReq, availableTransferChains)
	}

	chain := selector.SelectChain(chainReq, availableTransferChains)

	return chain
}
