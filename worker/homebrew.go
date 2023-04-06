package worker

import (
	"context"
	"errors"
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

const IndexStartDate = "2023-04-03"
const TimeFormat = "2006-01-02"

type homebrewDBClient struct {
	Container *azcosmos.ContainerClient
	OsType    database.OsType
	cache     map[int]homebrewcaculator.CountInfo
}

func hewHomebrewDBClient(container *azcosmos.ContainerClient, osType database.OsType) homebrewDBClient {
	return homebrewDBClient{
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

	err = database.CreateOrUpdateItem(ctx, h.Container, string(h.OsType), dbObj)
	if err != nil {
		return err
	}

	h.cache[idx] = data

	return nil
}

type HomebrewWorker struct {
	Container *azcosmos.ContainerClient
	OsTypes   []database.OsType
}

func (w HomebrewWorker) Run(ctx context.Context, date string) error {
	log.Println("[Homebrew] fetch data")
	apiFailure := false
	hbResp, err := datasource.FetchHomeBrewDownloadCount(date)
	if err != nil {
		log.Printf("[Homebrew] fetch homebrew data failed: %+v\r\n", err)
		apiFailure = true
	}

	log.Println("[Homebrew] write raw data to db")
	for _, item := range hbResp {
		item.Id = newHomebrewItemId(date, item.OsType)
		item.ApiFailure = apiFailure
		err := database.CreateOrUpdateItem(ctx, w.Container, item.OsType, item)
		if err != nil {
			return err
		}
	}

	log.Println("[Homebrew] begin calc")
	var errs error
	for _, osType := range w.OsTypes {
		var calcDBClient homebrewcaculator.DatabaseClient = hewHomebrewDBClient(w.Container, osType)

		calculator := homebrewcaculator.NewCalculator([]homebrewcaculator.Span{ThirtyDaysSpan, NinetyDaysSpan, OneYearSpan}, &calcDBClient)
		err := calculator.Calc(ctx, dateStr2Idx(date))
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
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
