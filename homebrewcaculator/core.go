package homebrewcaculator

type mockDBClient struct{}

func (m mockDBClient) Get(int) (CountInfo, error) {
	return CountInfo{}, nil
}

func (m mockDBClient) Update(idx int, data CountInfo) error {
	return nil
}
