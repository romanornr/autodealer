// Copyright (c) 2021 Romano
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package engineManager

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/romanornr/autodealer/flagparser"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/engine"
	gctlog "github.com/thrasher-corp/gocryptotrader/log"
)

func StartMainEngine() tea.Msg {
	var err error
	settings, flagSet := flagparser.DefaultEngineSettings()

	engine.Bot, err = engine.NewFromSettings(&settings, flagSet)
	if engine.Bot == nil || err != nil {
		logrus.Fatalf("Unable to initialise bot engine. Error: %s\n", err)
	}

	if err = engine.Bot.Start(); err != nil {
		gctlog.Errorf(gctlog.Global, "Unable to start bot engine. Error: %s\n", err)
		logrus.Errorf("Unable to start bot engine. Error: %s\n", err)
		os.Exit(1)
	}
	return SuccessMsg{Msg: "Main engine successfully started"}
}

func StopMainEngine() tea.Cmd {
	m := NewModel()
	engine.Bot.Stop()
	_, cmd := m.Update(SuccessMsg{Msg: "Main engine stopped successfully"})
	return cmd
}

func SpanNewEngine() (*engine.Engine, error) {
	settings, flagSet := flagparser.DefaultEngineSettings()
	e, err := engine.NewFromSettings(&settings, flagSet)
	if e == nil || err != nil {
		logrus.Warnf("Unable to initialise engine. Error: %s\n", err)
	}
	return e, err
}
