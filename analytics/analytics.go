package analytics

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-echarts/go-echarts/charts"
	"github.com/jinzhu/now"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/engine"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"gonum.org/v1/gonum/stat"
)

// TODO IGNORE FOR NOW

func GetSummary(pair currency.Pair, outputChart bool, forceBTC bool) {
	// defer timeTrack(time.Now(), "sum")

	exchange, _ := engine.Bot.GetExchangeByName("Binance")
	var err error
	if forceBTC == true {
		pair = currency.NewPair(currency.BTC, currency.USDT)
		exchange.SetPairs(currency.Pairs{pair}, asset.Spot, true)
	}

	// dailyOpen := time.Now().Add(-24 * time.Hour)
	// yesterdayOpen := time.Date(dailyOpen.Year(), dailyOpen.Month(), dailyOpen.Day(), 1, 0, 0, 0, dailyOpen.Location())

	klines, err := exchange.GetHistoricCandles(pair, asset.Spot, now.BeginningOfWeek(), now.EndOfDay(), kline.FifteenMin)
	if err != nil {
		logrus.Warnf("Failed getting historical candles: %s:%s\n", pair.String(), err)
	}

	var closes []float64

	x := make([]string, 0)
	xSlope := make([]float64, 0)
	y := make([][]float64, 0)
	ySlope := make([]float64, 0)

	var lastCandleClosingPrice = 0.0
	for i, k := range klines.Candles {
		closes = append(closes, k.Close)
		hour, minute, _ := k.Time.UTC().Clock()
		x = append(x, fmt.Sprintf("%s:%s", strconv.Itoa(hour), strconv.Itoa(minute)))
		xSlope = append(xSlope, float64(i))
		y = append(y, []float64{k.Open, k.Close, k.Low, k.High})
		ySlope = append(ySlope, k.Close)
		lastCandleClosingPrice = k.Close
	}

	if outputChart == true && len(klines.Candles) > 1 {
		go renderChart(klines, x, y)
	}

	alpha, beta := stat.LinearRegression(xSlope, ySlope, nil, false)
	r2 := stat.RSquared(xSlope, ySlope, nil, alpha, beta)

	logrus.Infof("start: %s\n", klines.Candles[0].Time.String())
	logrus.Infof("end: %s\n\n", klines.Candles[len(klines.Candles)-1].Time.String())

	logrus.Infof("Estimated slope is:  %.6f\n", alpha)
	logrus.Infof("Estimated offset is: %.6f\n", beta)
	logrus.Infof("R^2: %.6f\n", r2)

	mean, stdev := stat.MeanStdDev(closes, nil)
	zscore := stat.StdScore(lastCandleClosingPrice-mean/stdev, mean, stdev)
	logrus.Infof("The mean is: %f\n", mean)
	logrus.Infof("Standard deviations: %f\n", stdev)
	logrus.Infof("Z-score: %f\n", zscore)
}

func renderChart(klines kline.Item, x []string, y [][]float64) {
	chart := charts.NewKLine()
	chart.AddXAxis(x).AddYAxis("kline", y)
	chart.SetGlobalOptions(
		charts.TitleOpts{Title: klines.Pair.String()},
		charts.XAxisOpts{SplitNumber: 20},
		charts.YAxisOpts{Scale: true},
		charts.DataZoomOpts{XAxisIndex: []int{0}, Start: 0, End: 100},
		charts.InitOpts{Theme: "white"},
	)
	f, err := os.Create("kline.html")
	if err != nil {
		logrus.Warnf("failed to render chart %s %s: %s\n", klines.Pair, klines.Exchange, err)
	}
	chart.Render(f)
}
