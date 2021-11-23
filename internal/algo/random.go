package algo

import (
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
)

func RandomizeSize(minimalSize, maximumSize float64) decimal.Decimal {
	rand.Seed(time.Now().Unix())
	x := rand.Float64()*(maximumSize-minimalSize) + minimalSize
	r := decimal.NewFromFloat(x)
	hundred := decimal.New(100, 1)
	return r.Mul(hundred).Div(hundred)
}

func RandFloats(min, max float64, n int) []float64 {
	res := make([]float64, n)
	for i := range res {
		res[i] = min + rand.Float64()*(max-min)
	}
	return res
}
