package homebrewcaculator_test

import (
	"context"
	"log"
	"reflect"
	"testing"

	"aztfy-download-counter/homebrewcaculator"
)

type MockDBClient struct {
	MockData map[int]homebrewcaculator.CountInfo
}

func (m MockDBClient) Get(_ context.Context, idx int) (homebrewcaculator.CountInfo, error) {
	if idx < 0 {
		return homebrewcaculator.CountInfo{
			Count: 0,
		}, nil
	}
	if idx >= len(m.MockData) {
		return homebrewcaculator.CountInfo{
			Count: -1,
		}, nil
	}
	return m.MockData[idx], nil
}

func (m MockDBClient) Set(_ context.Context, idx int, data homebrewcaculator.CountInfo) error {
	m.MockData[idx] = data
	return nil
}

type testCase struct {
	spans    []homebrewcaculator.Span
	data     map[int]homebrewcaculator.CountInfo
	expect   map[int]homebrewcaculator.CountInfo
	startIdx int
}

type testingWriter struct {
	t *testing.T
}

func (w *testingWriter) Write(p []byte) (n int, err error) {
	w.t.Log(string(p))
	return len(p), nil
}

func TestCalculator_Calc(t *testing.T) {
	cases := map[string]testCase{
		"count right":       case_todayCountRight(),
		"count left":        case_todayCountLeft(),
		"total count right": case_todayTotalCountRight(),
		"total count left":  case_todayTotalCountLeft(),
		"count recursion":   case_CountRecursion(),
	}

	for name, c := range cases {
		t.Run(
			name, func(t *testing.T) {
				ctx := context.TODO()
				var mockDBClient homebrewcaculator.DatabaseClient = MockDBClient{
					MockData: c.data,
				}

				calculator := homebrewcaculator.NewCalculator(c.spans, &mockDBClient, log.New(&testingWriter{t}, "", log.LstdFlags))
				err := calculator.Calc(ctx, c.startIdx)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if !reflect.DeepEqual(c.expect, c.data) {
					t.Fatalf("result does not match!\rexpect:\r%v\rgot\r%v", c.expect, c.data)
				}
			})
	}
}

func case_todayCountRight() testCase {
	foo := homebrewcaculator.CountInfo{
		Count: 0,
		TotalCounts: map[homebrewcaculator.Span]int{
			2: 0,
		},
	}

	return testCase{
		spans: []homebrewcaculator.Span{2},
		data: map[int]homebrewcaculator.CountInfo{
			0: foo,
			1: foo,
			2: foo,
			3: {
				Count: -1,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 3,
				},
			},
		},
		expect: map[int]homebrewcaculator.CountInfo{
			0: foo,
			1: foo,
			2: foo,
			3: {
				Count: 3,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 3,
				},
			},
		},
		startIdx: 3,
	}
}

func case_todayCountLeft() testCase {
	foo := homebrewcaculator.CountInfo{
		Count: 0,
		TotalCounts: map[homebrewcaculator.Span]int{
			2: 0,
			5: 0,
			8: 0,
		},
	}

	return testCase{
		spans: []homebrewcaculator.Span{2},
		data: map[int]homebrewcaculator.CountInfo{
			0: foo,
			1: foo,
			2: {
				Count:       -1,
				TotalCounts: map[homebrewcaculator.Span]int{},
			},
			3: {
				Count: 2,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 3,
				},
			},
		},
		expect: map[int]homebrewcaculator.CountInfo{
			0: foo,
			1: foo,
			2: {
				Count: 1,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 1,
				},
			},
			3: {
				Count: 2,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 3,
				},
			},
		},
		startIdx: 3,
	}
}

func case_todayTotalCountRight() testCase {
	foo := homebrewcaculator.CountInfo{
		Count: 0,
		TotalCounts: map[homebrewcaculator.Span]int{
			2: 0,
		},
	}

	return testCase{
		spans: []homebrewcaculator.Span{2},
		data: map[int]homebrewcaculator.CountInfo{
			0: foo,
			1: foo,
			2: {
				Count: 1,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 1,
				},
			},
			3: {
				Count:       2,
				TotalCounts: map[homebrewcaculator.Span]int{},
			},
		},
		expect: map[int]homebrewcaculator.CountInfo{
			0: foo,
			1: foo,
			2: {
				Count: 1,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 1,
				},
			},
			3: {
				Count: 2,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 3,
				},
			},
		},
		startIdx: 3,
	}
}

func case_todayTotalCountLeft() testCase {
	foo := homebrewcaculator.CountInfo{
		Count: 0,
		TotalCounts: map[homebrewcaculator.Span]int{
			2: 0,
		},
	}

	return testCase{
		spans: []homebrewcaculator.Span{2},
		data: map[int]homebrewcaculator.CountInfo{
			0: foo,
			1: foo,
			2: {
				Count:       1,
				TotalCounts: map[homebrewcaculator.Span]int{},
			},
			3: {
				Count: 2,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 3,
				},
			},
		},
		expect: map[int]homebrewcaculator.CountInfo{
			0: foo,
			1: foo,
			2: {
				Count: 1,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 1,
				},
			},
			3: {
				Count: 2,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 3,
				},
			},
		},
		startIdx: 3,
	}
}

func case_CountRecursion() testCase {
	foo := homebrewcaculator.CountInfo{
		Count: 0,
		TotalCounts: map[homebrewcaculator.Span]int{
			2: 0,
			5: 0,
		},
	}

	return testCase{
		spans: []homebrewcaculator.Span{2, 5},
		data: map[int]homebrewcaculator.CountInfo{
			0: foo,
			1: foo,
			2: {
				Count:       -1,
				TotalCounts: map[homebrewcaculator.Span]int{},
			},
			3: {
				Count: 0,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 1,
					5: 1,
				},
			},
			4: foo,
			5: foo,
			6: {
				Count: 0,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 0,
					5: 1,
				},
			},
			7: {
				Count: -1,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 2,
					5: 2,
				},
			},
		},
		expect: map[int]homebrewcaculator.CountInfo{
			0: foo,
			1: foo,
			2: {
				Count: 1,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 1,
					5: 1,
				},
			},
			3: {
				Count: 0,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 1,
					5: 1,
				},
			},
			4: foo,
			5: foo,
			6: {
				Count: 0,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 0,
					5: 1,
				},
			},
			7: {
				Count: 2,
				TotalCounts: map[homebrewcaculator.Span]int{
					2: 2,
					5: 2,
				},
			},
		},
		startIdx: 7,
	}
}
