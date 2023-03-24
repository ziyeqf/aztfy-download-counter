package homebrewcaculator

type mockDBClient struct{}

func (m mockDBClient) IsAValidDate(date Date) bool {
	return true
}

func (m mockDBClient) GetDataAtDayN(date Date) (Data, error) {
	return Data{}, nil
}

func (m mockDBClient) UpdateDataAtDayN(date Date, data Data) error {
	return nil
}

func Start(startDate Date) {
	client := mockDBClient{}

	queue := Queue{}
	for _, span := range PossibleSpans() {
		queue.Enqueue(calculateDayCountAtDayN(startDate, span))
		queue.Enqueue(calculateHistoryCountAtDayN(startDate, span))
		queue.Enqueue(calculateHistoryCountAtDayNByNewerData(startDate, span))
		queue.Enqueue(calculateDayCountAtDayNByNewerData(startDate, span))
	}

	for !queue.IsEmpty() {
		calculateFuncPtr := queue.Dequeue()
		if calculateFuncPtr == nil {
			calculateFunc := *calculateFuncPtr
			newFuncs, err := calculateFunc(client)
			if err != nil {
				panic(err)
			}
			queue.Enqueue(newFuncs...)
		}
	}
}
