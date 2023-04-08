package referral

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ftx"
)

// Referral holds the referral information for a user
func GetFtxReferralLink() []ftx.ReferralRebateHistory {
	fmt.Println("check referral")
	cfg := config.GetConfig()
	err := cfg.LoadConfig("/home/romano/.gocryptotrader/config.json", true)
	if err != nil {
		logrus.Error("FTX load config error", err)
	}

	ftxConnfig, err := cfg.GetExchangeConfig("FTX")
	if err != nil {
		logrus.Errorf("GetExchangeConfig: %v", err)
		return nil
	}

	var f ftx.FTX
	f.SetDefaults()
	f.SetCredentials(ftxConnfig.API.Credentials.Key, ftxConnfig.API.Credentials.Secret, "", "", "", "")
	logrus.Printf("API key: %s", ftxConnfig.API.Credentials.Secret)
	err = f.Setup(ftxConnfig)
	if err != nil {
		logrus.Errorf("Setup: %v", err)
		return nil
	}
	///f.SetCredentials(apiKey, apiSecret, subAccount, "", "", "")

	r, err := f.GetReferralRebateHistory(context.Background())
	if err != nil {
		logrus.Print(err)
	}

	fmt.Println("referral rebates")
	fmt.Println(r)
	return r
}
