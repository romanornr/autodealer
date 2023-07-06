package webserver

//// MoveHandler handleMove is the handler for the '/move' page request.
//func MoveHandler(w http.ResponseWriter, _ *http.Request) {
//	line := charts.NewLine()
//	// set some global options like Title/Legend/ToolTip or anything else
//	line.SetGlobalOptions(
//		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
//		charts.WithYAxisOpts(opts.YAxis{Scale: true}),
//		charts.WithTitleOpts(opts.Title{
//			Title:    "FTX Move Contracts",
//			Subtitle: "TermStructure",
//		}))
//
//	d := singleton.GetDealer()
//	termStructure := move.GetTermStructure(d)
//
//	items := make([]opts.LineData, 0)
//	yesterday := make([]opts.LineData, 0)
//	var xstring []string
//
//	for _, m := range termStructure.MOVE.Statistic {
//		items = append(items, opts.LineData{Value: m.Stats.Greeks.ImpliedVolatility})
//		xstring = append(xstring, m.Data.ExpiryDescription)
//
//		//	yesterday = append(yesterday, opts.LineData{Value: m.Mark + m.Change24h})
//		//	xstring = append(xstring, m.ExpiryDescription)
//		//	//line.SetXAxis([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
//	}
//
//	// remove first element because the "Today" MOVE contract does not belong in the term structure
//	items = items[1:]
//
//	line.SetXAxis(xstring).
//		AddSeries("move", items).
//		AddSeries("yesterday", yesterday).
//		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
//
//	line.Render(w)
//}

//r.Route(routeMoveTermStructure, func(r chi.Router) {
//	r.Use(MoveTermStructureCtx)
//	r.Get("/", getMoveTermStructure)
//})
//
//r.Route(routeMoveStats, func(r chi.Router) {
//	r.Use(MoveStatsCtx)
//	r.Get("/", getMoveStats)
//})
