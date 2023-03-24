package homebrewcaculator

type CalculateDataAtDayNFunc func(client DatabaseClient) ([]CalculateDataAtDayNFunc, error)

// DayCount -> ed(n)
// HistoryCount -> xd(n)

// ed(n) = xd(n) - xd(n-1) + d(n-x)
func calculateDayCountAtDayN(date Date, span Span) CalculateDataAtDayNFunc {
	return func(client DatabaseClient) ([]CalculateDataAtDayNFunc, error) {
		todayData, err := client.GetDataAtDayN(date)
		if err != nil {
			return nil, err
		}
		yesterdayData, err := client.GetDataAtDayN(date.AddDays(-1))
		if err != nil {
			return nil, err
		}
		previousData, err := client.GetDataAtDayN(date.AddDays(-int(span)))
		if err != nil {
			return nil, err
		}

		_, todayDataOK := todayData.PreviousData[span]
		_, yesterdayDataOK := yesterdayData.PreviousData[span]
		previousDataOK := todayData.TodayCount != -1

		if !todayDataOK || !yesterdayDataOK || !previousDataOK {
			return nil, nil
		}

		todayData.TodayCount = todayData.PreviousData[span] - yesterdayData.PreviousData[span] + previousData.TodayCount
		err = client.UpdateDataAtDayN(date, todayData)
		if err != nil {
			return nil, err
		}

		return AfterCalculateDayCountAtDayN(date), nil
	}
}

func AfterCalculateDayCountAtDayN(date Date) []CalculateDataAtDayNFunc {
	var result []CalculateDataAtDayNFunc

	for _, span := range PossibleSpans() {
		result = append(result, calculateDayCountAtDayN(date.AddDays(int(span)), span))
		result = append(result, calculateHistoryCountAtDayN(date.AddDays(int(span)), span))
		result = append(result, calculateHistoryCountAtDayNByNewerData(date.AddDays(int(span)-1), span))
		result = append(result, calculateDayCountAtDayNByNewerData(date.AddDays(-int(span)), span)) // for other spans
	}

	return result
}

// xd(n) = ed(n) + xd(n-1) - d(n-x)
func calculateHistoryCountAtDayN(date Date, span Span) CalculateDataAtDayNFunc {
	return func(client DatabaseClient) ([]CalculateDataAtDayNFunc, error) {
		todayData, err := client.GetDataAtDayN(date)
		if err != nil {
			return nil, err
		}
		yesterdayData, err := client.GetDataAtDayN(date.AddDays(-1))
		if err != nil {
			return nil, err
		}
		previousData, err := client.GetDataAtDayN(date.AddDays(-int(span)))
		if err != nil {
			return nil, err
		}

		todayDataOk := todayData.TodayCount != -1
		_, yesterdatDataOK := yesterdayData.PreviousData[span]
		previousDataOK := previousData.TodayCount != -1

		if !todayDataOk || !yesterdatDataOK || !previousDataOK {
			return nil, nil
		}

		todayData.PreviousData[span] = todayData.TodayCount + yesterdayData.PreviousData[span] - previousData.TodayCount
		err = client.UpdateDataAtDayN(date, todayData)
		if err != nil {
			return nil, err
		}

		return AfterCalculateHistoryCountAtDayN(date), nil
	}
}

func AfterCalculateHistoryCountAtDayN(date Date) []CalculateDataAtDayNFunc {
	var result []CalculateDataAtDayNFunc

	for _, span := range PossibleSpans() {
		result = append(result, calculateDayCountAtDayN(date.AddDays(1), span))
		result = append(result, calculateHistoryCountAtDayN(date.AddDays(1), span))
		result = append(result, calculateDayCountAtDayNByNewerData(date.AddDays(1-int(span)), span))
		result = append(result, calculateDayCountAtDayNByNewerData(date.AddDays(-int(span)), span)) // for the other spans, there is a might to figure out.
	}

	return result
}

// d(n-x) = xd(n) - xd(n-1) + ed(n)
func calculateDayCountAtDayNByNewerData(date Date, span Span) CalculateDataAtDayNFunc {
	return func(client DatabaseClient) ([]CalculateDataAtDayNFunc, error) {
		todayData, err := client.GetDataAtDayN(date)
		if err != nil {
			return nil, err
		}
		futureData, err := client.GetDataAtDayN(date.AddDays(int(span)))
		if err != nil {
			return nil, err
		}
		previousData, err := client.GetDataAtDayN(date.AddDays(int(span) - 1))
		if err != nil {
			return nil, err
		}

		_, futureDataOk := futureData.PreviousData[span]
		_, previousDataOk := previousData.PreviousData[span]
		futureDataOk = futureDataOk && futureData.TodayCount != -1

		if !futureDataOk || !previousDataOk {
			return nil, nil
		}

		previousData.TodayCount = futureData.PreviousData[span] - previousData.PreviousData[span] + futureData.TodayCount
		err = client.UpdateDataAtDayN(date, todayData)
		if err != nil {
			return nil, err
		}

		return AfterCalculateDayCountAtDayNByNewerData(date), nil
	}
}
func AfterCalculateDayCountAtDayNByNewerData(date Date) []CalculateDataAtDayNFunc {
	var result []CalculateDataAtDayNFunc

	for _, span := range PossibleSpans() {
		result = append(result, calculateHistoryCountAtDayN(date, span))
		result = append(result, calculateHistoryCountAtDayNByNewerData(date.AddDays(-1), span))
		result = append(result, calculateDayCountAtDayNByNewerData(date.AddDays(-int(span)), span))
		result = append(result, calculateDayCountAtDayN(date.AddDays(int(span)), span))                  // for the other spans, there is a might to figure out.
		result = append(result, calculateHistoryCountAtDayN(date.AddDays(int(span)), span))              // for the other spans, there is a might to figure out.
		result = append(result, calculateHistoryCountAtDayNByNewerData(date.AddDays(int(span)-1), span)) // for the other spans, there is a might to figure out.
	}

	return result
}

// xd(n-1) = xd(n) - ed(n) + d(n-x)
func calculateHistoryCountAtDayNByNewerData(date Date, span Span) CalculateDataAtDayNFunc {
	return func(client DatabaseClient) ([]CalculateDataAtDayNFunc, error) {
		todayData, err := client.GetDataAtDayN(date)
		if err != nil {
			return nil, err
		}
		tommorrowData, err := client.GetDataAtDayN(date.AddDays(1))
		if err != nil {
			return nil, err
		}
		previousData, err := client.GetDataAtDayN(date.AddDays(-int(span)))
		if err != nil {
			return nil, err
		}

		_, tomorrowDataOk := tommorrowData.PreviousData[span]
		previousDataOk := previousData.TodayCount != -1
		tomorrowDataOk = tomorrowDataOk && tommorrowData.TodayCount != -1

		if !tomorrowDataOk || !previousDataOk {
			return nil, nil
		}
		todayData.PreviousData[span] = tommorrowData.PreviousData[span] - tommorrowData.TodayCount + previousData.TodayCount
		err = client.UpdateDataAtDayN(date, todayData)
		if err != nil {
			return nil, err
		}

		return AfterCalculateHistoryCountAtDayNByNewerData(date), nil
	}
}

func AfterCalculateHistoryCountAtDayNByNewerData(date Date) []CalculateDataAtDayNFunc {
	var result []CalculateDataAtDayNFunc

	for _, span := range PossibleSpans() {
		result = append(result, calculateDayCountAtDayN(date, span))
		result = append(result, calculateHistoryCountAtDayN(date.AddDays(-1), span))
		result = append(result, calculateDayCountAtDayNByNewerData(date.AddDays(-int(span)), span))
		result = append(result, calculateHistoryCountAtDayNByNewerData(date.AddDays(1-int(span)), span)) // for other spans with same pattern
	}

	return result
}
