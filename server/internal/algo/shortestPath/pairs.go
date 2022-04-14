package shortestPath

import (
	"context"
	"fmt"
	"github.com/RyanCarrier/dijkstra"
	"github.com/romanornr/autodealer/internal/singleton"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

// MatchPairsForCurrency returns a list of pairs that match the given currency
func MatchPairsForCurrency(e exchange.IBotExchange, code currency.Code, assetType asset.Item) currency.Pairs {
	availablePairs, err := e.GetAvailablePairs(assetType)
	if err != nil {
		return nil
	}

	matchingPairs := currency.Pairs{}
	for _, pair := range availablePairs {
		if pair.Base == code {
			matchingPairs = append(matchingPairs, pair)
		}
	}

	return matchingPairs
}

type Nodes struct {
	Nodes []*Node
}

type Node struct {
	Pairs currency.Pairs
	ID    int
}

// PathToAsset returns the shortest path to the given asset
func PathToAsset(e exchange.IBotExchange, code currency.Code, destination currency.Code, assetType asset.Item) ([]currency.Code, error) {

	availablePairs := currency.Pairs{}
	// chosen asset is "ANC"
	// INFO[0121]: relatablePairs [ANC-BTC ANC-BUSD ANC-USDT]
	relatablePairs := MatchPairsForCurrency(e, code, assetType)

	availablePairs = append(availablePairs, relatablePairs...)

	nodes := Nodes{}
	nodes.Nodes = append(nodes.Nodes, &Node{Pairs: availablePairs, ID: 0})

	graph := dijkstra.NewGraph()

	for i, p := range availablePairs {
		m := MatchPairsForCurrency(e, p.Quote, assetType)

		nodes.Nodes = append(nodes.Nodes, &Node{Pairs: m, ID: i + 1})
		availablePairs = append(availablePairs, m...)
	}

	vertices := make(map[currency.Code]int)

	count := 0
	for _, n := range nodes.Nodes {
		// each n.Pairs.Quote is a currency.Code and each n.Pairs.Base is a currency.Code
		for _, p := range n.Pairs {
			vertices[p.Quote] = int(count)
			count++
			vertices[p.Base] = int(count)
			count++

			graph.AddVertex(vertices[p.Quote])
			graph.AddVertex(vertices[p.Base])
		}

		//INFO[0121] 1: [BTC-USDT BTC-TUSD BTC-USDC BTC-BUSD BTC-NGN BTC-RUB BTC-TRY BTC-EUR BTC-GBP BTC-UAH BTC-BIDR BTC-AUD BTC-DAI BTC-BRL BTC-USDP]
		//INFO[0121] 2: [BUSD-USDT BUSD-RUB BUSD-TRY BUSD-BIDR BUSD-DAI BUSD-BRL BUSD-VAI BUSD-UAH]
		//INFO[0121] 3: [USDT-TRY USDT-RUB USDT-IDRT USDT-UAH USDT-BIDR USDT-DAI USDT-NGN USDT-BRL]
		//logrus.Printf("%d: %s", n.ID, n.Pairs)
	}

	for _, n := range nodes.Nodes {
		// each n.Pairs.Quote is a currency.Code and each n.Pairs.Base is a currency.Code
		for _, p := range n.Pairs {
			err := graph.AddArc(vertices[p.Base], vertices[p.Quote], int64(n.ID))
			if err != nil {
				return nil, fmt.Errorf("error adding arc: %s", err)
			}
		}
	}

	// logrus.Printf("vertices %v\n", vertices)

	// add the edges to the graph
	best, err := graph.Shortest(vertices[code], vertices[destination])
	if err != nil {
		logrus.Errorf("error finding best path: %s", err)
	}

	fmt.Println("Shortest distance ", best.Distance, " following path ", best.Path)

	var codesToPath []currency.Code

	// convert the path to currency codes
	for _, n := range best.Path {
		key, ok := mapkeyVertices(vertices, n)
		if !ok {
			return nil, fmt.Errorf("error finding key for %d", n)
		}
		codesToPath = append(codesToPath, key)
	}

	return codesToPath, nil
}

// mapkeyVertices returns the key for the given value
func mapkeyVertices(m map[currency.Code]int, value int) (key currency.Code, ok bool) {
	for k, v := range m {
		if v == value {
			key = k
			ok = true
			return
		}
	}
	return
}

func fetchTickerPrice(e exchange.IBotExchange, currencies []currency.Code, assetType asset.Item) (float64, error) {

	// currencies is a slice
	// currencies VIA, BTC, USD (3 currency codes)
	// which should become VIA-BTC, BTC-USD  (2 pairs created from 3 currency codes)

	d := singleton.GetDealer()
	e, err := d.ExchangeManager.GetExchangeByName(e.GetName())
	if err != nil {
		return 0, err
	}

	for _, c := range currencies {
		logrus.Printf("currency %s\n", c)
	}

	pairs := currency.Pairs{}

	// create pairs from the currencies Slice
	for i, c := range currencies {
		if i == len(currencies)-1 {
			break
		}
		pair := currency.NewPair(c, currencies[i+1]) // the quote is the next currency code
		pairs = append(pairs, pair)
	}

	var accumulatedPrice float64

	for _, p := range pairs {
		price, err := e.FetchTicker(context.Background(), p, assetType)
		if err != nil {
			return 0, err
		}

		if accumulatedPrice == 0.0 {
			accumulatedPrice = price.Last
			continue
		}

		accumulatedPrice = accumulatedPrice * price.Last
	}
	return accumulatedPrice, nil
}

// Write Bellman Ford Algorithm to find the shortest path to a dollar pair by using an algorithm
func WriteBellmanFordAlgorithm(e exchange.IBotExchange, code currency.Code, assetType asset.Item) {

}

// Get the dollar value of the given asset. However, there might not be a direct conversion to USD so we need to use the exchange's conversion rate
// to get the value in USD. Possibly use an intermediate currency pair to convert to USD.
