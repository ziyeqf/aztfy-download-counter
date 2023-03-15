package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"aztfy-download-counter/datasource"
	"aztfy-download-counter/kustohelper"
	"github.com/Azure/azure-kusto-go/kusto"
)

const dbName = "aztfyDownloadCount"
const ghTable = "Github"
const ghRawTable = "GithubRaw"
const hbRawTable = "HomebrewRaw"
const hbTable = "Homebrew"

func main() {
	ctx := context.TODO()
	clientId := os.Getenv("ARM_CLIENT_ID")
	clientSecret := os.Getenv("ARM_CLIENT_SECRET")
	tenantId := os.Getenv("ARM_TENANT_ID")
	clusterURI := os.Getenv("KUSTO_CLUSTER_URI")

	kustoClient, err := kustohelper.AuthKusto(clusterURI, clientId, clientSecret, tenantId)
	if err != nil {
		logError(err)
	}

	err = fetchForGithub(ctx, kustoClient)
	if err != nil {
		logError(err)
	}

	err = fetchForHomeBrew(ctx, kustoClient)
	if err != nil {
		logError(err)
	}
}

func fetchForGithub(ctx context.Context, kustoClient *kusto.Client) error {
	lastRunId, err := kustohelper.QueryRunIdForGithubRaw(ctx, kustoClient, dbName)
	if err != nil {
		return fmt.Errorf("query github raw data saved yesterday: %+v", err)
	}

	log.Println("fetch github data")
	ghResp, err := datasource.FetchGitHubDownloadCount(ctx, lastRunId+1)
	if err != nil {
		return err
	}

	_, err = saveGithubRaw(ctx, kustoClient, ghResp)
	if err != nil {
		return err
	}

	_, err = saveGithub(ctx, kustoClient, ghResp)
	if err != nil {
		return err
	}

	return err
}

func saveGithubRaw(ctx context.Context, kustoClient *kusto.Client, resp []interface{}) (chan error, error) {
	log.Println("save github raw data to kusto")
	githubRawQueryResp, err := kustohelper.QueryExistingRecordForGithubRaw(ctx, kustoClient, dbName, time.Now())
	if err != nil {
		return nil, fmt.Errorf("query github: %+v", err)
	}

	if len(githubRawQueryResp) > 0 {
		log.Println("github raw data already fetched today, skip this round")
		return nil, nil
	}

	return kustohelper.SaveToKusto(ctx, kustoClient, dbName, ghRawTable, kustohelper.GithubRawIngestMapping, resp)
}

func saveGithub(ctx context.Context, kustoClient *kusto.Client, resp []interface{}) (chan error, error) {
	githubQueryResp, err := kustohelper.QueryExistingRecordForGithub(ctx, kustoClient, dbName, time.Now())
	if err != nil {
		return nil, fmt.Errorf("query github: %s", err)
	}

	if len(githubQueryResp) > 0 {
		log.Println("github data already fetched today, skip this round")
		return nil, nil
	}

	newRows := make([]interface{}, 0)
	githubMap := make(map[string]datasource.GithubVersion)
	for _, item := range resp {
		gh := item.(datasource.GithubVersion)
		if _, ok := githubMap[gh.Ver]; !ok {
			githubMap[gh.Ver] = datasource.GithubVersion{
				Ver:         gh.Ver,
				CountDate:   time.Now(),
				PublishDate: gh.PublishDate,
				RunId:       gh.RunId,
			}
		}
		kItem := githubMap[gh.Ver]
		kItem.DownloadCount = kItem.DownloadCount + gh.DownloadCount
		githubMap[gh.Ver] = kItem
	}

	for k, v := range githubMap {
		log.Printf("[Github] ver: %s, count: %d", k, v.DownloadCount)
		newRows = append(newRows, v)
	}

	if len(newRows) > 0 {
		log.Println("save Github data to kusto")
		return kustohelper.SaveToKusto(ctx, kustoClient, dbName, ghTable, kustohelper.GithubIngestMapping, newRows)
	}

	return nil, nil
}

func fetchForHomeBrew(ctx context.Context, kustoClient *kusto.Client) error {
	lastRunId, err := kustohelper.QueryRunIdForHomeBrewRaw(ctx, kustoClient, dbName)
	if err != nil {
		return fmt.Errorf("query github raw data saved yesterday: %+v", err)
	}

	log.Println("fetch homebrew data")
	hbResp, err := datasource.FetchHomeBrewDownloadCount(lastRunId + 1)
	if err != nil {
		log.Println(err)
	}

	_, err = saveHomeBrewRaw(ctx, kustoClient, hbResp)
	if err != nil {
		return err
	}

	_, err = saveHomebrew(ctx, kustoClient, hbResp)
	if err != nil {
		return err
	}

	return nil
}

func saveHomeBrewRaw(ctx context.Context, kustoClient *kusto.Client, resp []interface{}) (chan error, error) {
	homebrewQueryResp, err := kustohelper.QueryExistingRecordForHomebrewRaw(ctx, kustoClient, dbName, time.Now())
	if err != nil {
		return nil, fmt.Errorf("query homebrew: %+v", err)
	}

	if len(homebrewQueryResp) > 0 {
		log.Println("homebrew raw data already fetched today, skip this round")
		return nil, nil
	}

	log.Println("save homebrew raw data to kusto")
	return kustohelper.SaveToKusto(ctx, kustoClient, dbName, hbRawTable, kustohelper.HomeBrewRawIngestMapping, resp)
}

func saveHomebrew(ctx context.Context, kustoClient *kusto.Client, resp []interface{}) (chan error, error) {
	homebrewQueryResp, err := kustohelper.QueryExistingRecordForHomebrew(ctx, kustoClient, dbName, time.Now())
	if err != nil {
		return nil, fmt.Errorf("query homebrew: %+v", err)
	}

	if len(homebrewQueryResp) > 0 {
		log.Println("homebrew data already fetched today, skip this round")
		return nil, nil
	}

	lastRunid, err := kustohelper.QueryRunIdForHomeBrew(ctx, kustoClient, dbName)

	newRows := make([]interface{}, 0)
	for _, osType := range []datasource.OsType{datasource.OsTypeLinux, datasource.OsTypeDarwin} { // homebrew only provide linux and macos data
		lastCount, err := kustohelper.QueryDownloadCountForHomebrew(ctx, kustoClient, dbName, time.Now().AddDate(0, 0, -30), osType)
		if err != nil {
			logError(err)
		}

		if lastCount == -1 {
			log.Printf("[HomeBrew] There is no count for %s saved on 30d ays ago, skip this round.\n", osType)
			continue
		}

		for _, hbRaw := range resp {
			hb := hbRaw.(datasource.HomeBrewVersion)
			if datasource.OsType(hb.OsType) != osType {
				continue
			}

			if datasource.HomeBrewDataType(hb.DataType) == datasource.ThirtyDays {
				newRows = append(newRows, datasource.HomeBrewVersion{
					OsType:        string(osType),
					CountDate:     time.Now(),
					DownloadCount: lastCount + hb.DownloadCount,
					RunId:         lastRunid + 1,
				})
			}
		}
	}

	if len(newRows) > 0 {
		log.Println("save homebrew data to kusto")
		return kustohelper.SaveToKusto(ctx, kustoClient, dbName, hbTable, kustohelper.HomeBrewIngestMapping, newRows)
	}

	return nil, nil
}

func logError(err error) {
	log.Println(err)
	os.Exit(1)
}
