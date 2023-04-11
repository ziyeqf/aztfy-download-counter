package job

import (
	"context"
	"fmt"
	"log"
	"time"

	"aztfy-download-counter/database"
	"aztfy-download-counter/datasource"
	"aztfy-download-counter/homebrewcaculator"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

const (
	ThirtyDaysSpan homebrewcaculator.Span = 30
	NinetyDaysSpan homebrewcaculator.Span = 90
	OneYearSpan    homebrewcaculator.Span = 365
)

const IndexStartDate = "2023-04-11"
const TimeFormat = "2006-01-02"

type homebrewDBClient struct {
	ContainerInitFunc func() (*azcosmos.ContainerClient, error)
	Logger            *log.Logger
	Container         *azcosmos.ContainerClient
	OsType            database.OsType
	cache             map[int]homebrewcaculator.CountInfo
}

func newHomebrewDBClient(container *azcosmos.ContainerClient, osType database.OsType, logger *log.Logger) homebrewDBClient {
	return homebrewDBClient{
		Logger:    logger,
		Container: container,
		OsType:    osType,
		cache:     make(map[int]homebrewcaculator.CountInfo),
	}
}

func (h homebrewDBClient) Get(ctx context.Context, idx int) (homebrewcaculator.CountInfo, error) {
	if idx < 0 {
		return homebrewcaculator.CountInfo{
			Count: 0,
			TotalCounts: map[homebrewcaculator.Span]int{
				ThirtyDaysSpan: 0,
				NinetyDaysSpan: 0,
				OneYearSpan:    0,
			},
		}, nil
	}

	if v, ok := h.cache[idx]; ok {
		return v, nil
	}

	date := idx2DateStr(idx)
	result := homebrewcaculator.CountInfo{
		Count:       -1,
		TotalCounts: make(map[homebrewcaculator.Span]int),
	}

	resp, err := database.QueryItem(ctx, h.Container, string(h.OsType), date, database.HomebrewVersion{})
	if len(resp) == 0 {
		return result, nil
	}

	if err == nil && len(resp) == 1 {
		result.Count = resp[0].TodayCount
	}

	// the unmarshal func will give them default values.
	// so we use `APIFailure` to determine if its Span data exists.
	if !resp[0].ApiFailure {
		result.TotalCounts = map[homebrewcaculator.Span]int{
			ThirtyDaysSpan: resp[0].ThirtyDayCount,
			NinetyDaysSpan: resp[0].NinetyDayCount,
			OneYearSpan:    resp[0].OneYearCount,
		}
	}
	return result, nil
}

func (h homebrewDBClient) Set(ctx context.Context, idx int, data homebrewcaculator.CountInfo) error {
	dbObjects, err := database.QueryItem(ctx, h.Container, string(h.OsType), idx2DateStr(idx), database.HomebrewVersion{})
	if err != nil {
		return err
	}
	if len(dbObjects) > 1 {
		return fmt.Errorf("get more than 1 item for %s, osType: %s", idx2DateStr(idx), h.OsType)
	}

	var dbObj database.HomebrewVersion
	if len(dbObjects) == 1 {
		dbObj = dbObjects[0]
	} else {
		dbObj = database.HomebrewVersion{
			Id:         newHomebrewItemId(idx2DateStr(idx), string(h.OsType)),
			OsType:     string(h.OsType),
			CountDate:  idx2DateStr(idx),
			ApiFailure: true,
		}
	}

	dbObj.TodayCount = data.Count
	for k, v := range data.TotalCounts {
		switch k {
		case ThirtyDaysSpan:
			dbObj.ThirtyDayCount = v
		case NinetyDaysSpan:
			dbObj.NinetyDayCount = v
		case OneYearSpan:
			dbObj.OneYearCount = v
		}
	}

	h.Logger.Printf("update homebrew data to db: %+v\r", dbObj)
	err = database.CreateOrUpdateItem(ctx, h.Container, string(h.OsType), dbObj)
	if err != nil {
		return err
	}

	h.cache[idx] = data

	return nil
}

type HomebrewWorker struct {
	Logger            *log.Logger
	ContainerInitFunc func() (container *azcosmos.ContainerClient, err error)
	OsTypes           []database.OsType
	Date              string
}

func (w HomebrewWorker) Run(ctx context.Context) {
	container, err := w.ContainerInitFunc()
	if err != nil {
		w.Logger.Println(err)
		return
	}

	w.Logger.Println("fetch data")
	apiFailure := false
	hbResp, err := datasource.FetchHomeBrewDownloadCount()
	if err != nil {
		w.Logger.Println("fetch homebrew data failed: %+v\r", err)
		apiFailure = true
	}

	var brewVersions []database.HomebrewVersion
	for _, osType := range []database.OsType{database.OsTypeDarwin, database.OsTypeLinux} {
		brewVersions = append(brewVersions, w.generateHomeBrewVersion(*hbResp, osType, w.Date, apiFailure))
	}

	w.Logger.Println("write raw data to db")
	for _, item := range brewVersions {
		err := database.CreateOrUpdateItem(ctx, container, item.OsType, item)
		if err != nil {
			w.Logger.Println(err)
			return
		}
	}

	w.Logger.Println("begin calc")
	for _, osType := range w.OsTypes {
		var calcDBClient homebrewcaculator.DatabaseClient = newHomebrewDBClient(container, osType, w.Logger)
		calcLogger := log.New(w.Logger.Writer(), w.Logger.Prefix()+"[Calc] ", 0)
		calculator := homebrewcaculator.NewCalculator([]homebrewcaculator.Span{ThirtyDaysSpan, NinetyDaysSpan, OneYearSpan}, &calcDBClient, calcLogger)
		err := calculator.Calc(ctx, dateStr2Idx(w.Date))
		if err != nil {
			w.Logger.Println(err)
		}
	}

	w.Logger.Println("done")
	return
}

func (w HomebrewWorker) generateHomeBrewVersion(input datasource.BrewJson, osType database.OsType, date string, apiFailure bool) database.HomebrewVersion {
	var i datasource.Install
	if osType == database.OsTypeDarwin {
		i = input.Analytics.Install
	} else {
		i = input.AnalyticsLinux.Install
	}

	output := database.HomebrewVersion{
		Id:             newHomebrewItemId(date, string(osType)),
		OsType:         string(osType),
		CountDate:      date,
		ThirtyDayCount: i.ThirtyDays.Aztfy,
		NinetyDayCount: i.NinetyDays.Aztfy,
		OneYearCount:   i.OneYear.Aztfy,
		ApiFailure:     apiFailure,
	}

	return output
}

func idx2DateStr(idx int) string {
	idxStart, _ := time.Parse(TimeFormat, IndexStartDate)
	return idxStart.AddDate(0, 0, idx).Format(TimeFormat)
}

func dateStr2Idx(date string) int {
	idxStart, _ := time.Parse(TimeFormat, IndexStartDate)
	dateTime, _ := time.Parse(TimeFormat, date)
	return int(dateTime.Sub(idxStart).Hours() / 24)
}

func newHomebrewItemId(date string, osType string) string {
	return fmt.Sprintf("%s-%s", date, osType)
}
