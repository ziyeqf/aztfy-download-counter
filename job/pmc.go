package job

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"aztfy-download-counter/database"
	"aztfy-download-counter/datasource"
	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

const startDate = "2022-10-21"

type PMCWorker struct {
	ContainerInitFunc func() (*azcosmos.ContainerClient, error)
	Logger            *log.Logger
	KustoEndpoint     string
	Date              string
	ArmClientId       string
	ArmClientSecret   string
	ArmTenantId       string
}

func (w PMCWorker) Run(ctx context.Context) {
	container, err := w.ContainerInitFunc()
	if err != nil {
		w.Logger.Println(err)
		return
	}

	kustoClient, err := datasource.AuthKusto(w.ArmClientId, w.ArmClientSecret, w.ArmTenantId, w.KustoEndpoint)
	if err != nil {
		w.Logger.Println(fmt.Errorf("auth kusto failed, skipped: %v", err))
		return
	}
	defer func(kustoClient *kusto.Client) {
		err := kustoClient.Close()
		if err != nil {
			w.Logger.Println(err)
			return
		}
	}(kustoClient)

	datetime, err := time.Parse(TimeFormat, w.Date)
	if err != nil {
		w.Logger.Println(err)
		return
	}

	resp, err := datasource.QueryForPMC(ctx, kustoClient, datetime)
	if err != nil {
		w.Logger.Println(err)
		return
	}

	// [version][arch]PMCVersion
	result := make(map[string]map[string]*database.PMCVersion)
	for _, item := range resp {
		version, arch, err := w.parseTagNameForRPM(item.Path)
		if err != nil {
			w.Logger.Printf("parseTagNameForRPM path: %v, error: %+v\r", item.Path, err)
		}
		if _, ok := result[version]; !ok {
			result[version] = make(map[string]*database.PMCVersion)
		}
		if _, ok := result[version][arch]; !ok {
			result[version][arch] = &database.PMCVersion{
				Id:         w.newPMCItemId(w.Date, arch, version),
				Date:       w.Date,
				Ver:        version,
				Arch:       arch,
				TodayCount: 0,
			}
		}
		result[version][arch].TodayCount++
	}

	// calculate totalCount
	for _, m := range result {
		for arch, item := range m {
			prevTotalCount, err := w.getPrevTotalCount(ctx, container, kustoClient, arch, item.Ver)
			if err != nil {
				w.Logger.Println(fmt.Errorf("getting prevTotalCount failed, skipped: %v", err))
				continue
			}
			item.TotalCount = prevTotalCount + int64(item.TodayCount)
		}
	}

	// [arch]PMCVersion
	dbObjMap := make(map[string][]database.PMCVersion)
	for _, m := range result {
		for arch, item := range m {
			dbObjMap[arch] = append(dbObjMap[arch], *item)
		}
	}

	w.Logger.Println("write PMC data to db")
	for arch, array := range dbObjMap {
		err = database.BatchUpsert(ctx, container, arch, array)
		if err != nil {
			w.Logger.Println(err)
			return
		}
	}

	w.Logger.Println("done")
}

func (w PMCWorker) getPrevTotalCount(ctx context.Context, container *azcosmos.ContainerClient, kustoClient *kusto.Client, arch string, version string) (int64, error) {
	d, err := time.Parse(TimeFormat, w.Date)
	if err != nil {
		return 0, err
	}
	itemId := w.newPMCItemId(d.AddDate(0, 0, -1).Format(TimeFormat), arch, version)

	prevObj := database.PMCVersion{}
	err = database.ReadItem(ctx, container, arch, itemId, &prevObj)
	if err != nil {
		if !strings.Contains(err.Error(), "NotFound") {
			return 0, err
		}
		w.Logger.Printf("there is no data with id %s, query from pmc data base", itemId)
		s, _ := time.Parse(TimeFormat, startDate)
		cnt, err := datasource.QueryTotalCount(ctx, kustoClient, s, d, version, arch)
		if err != nil {
			w.Logger.Println(err)
			return 0, err
		}
		return cnt, nil
	}

	return prevObj.TotalCount, nil
}

func (w PMCWorker) parseTagNameForRPM(tagName string) (version string, arch string, err error) {
	reg := regexp.MustCompile(`.*-(\d*\.\d*\.\d*)(-1-){0,1}(.+)\.rpm`)
	result := reg.FindStringSubmatch(tagName)
	if len(result) != 4 {
		return "", "", fmt.Errorf("parse failed")
	}
	return result[1], result[3], nil
}

func (w PMCWorker) newPMCItemId(date string, arch string, version string) string {
	return fmt.Sprintf("%s-%s-%s", date, arch, version)
}
