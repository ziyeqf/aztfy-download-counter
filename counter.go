package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"aztfy-download-counter/datasource"
	"aztfy-download-counter/kustohelper"
)

const dbName = "aztfyDownloadCount"
const ghTable = "Github"
const hbRawTable = "HomebrewRaw"
const hbTable = "HomeBrew"

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

	log.Println("fetch github data")
	ghResp, err := datasource.FetchGitHubDownloadCount(ctx)
	if err != nil {
		log.Println(err)
	}

	log.Println("save github data to kusto")
	ch, err := kustohelper.SaveToKusto(ctx, kustoClient, dbName, ghTable, kustohelper.GithubIngestMapping, ghResp)
	if err != nil {
		logError(fmt.Errorf("github: %s", err))
	}

	<-ch

	log.Println("fetch homebrew data")
	hbResp, err := datasource.FetchHomeBrewDownloadCount()
	if err != nil {
		log.Println(err)
	}

	log.Println("save homebrew raw data to kusto")
	ch, err = kustohelper.SaveToKusto(ctx, kustoClient, dbName, hbRawTable, kustohelper.HomeBrewRawIngestMapping, hbResp)
	if err != nil {
		logError(fmt.Errorf("HomeBrew:%s", err))
	}
	<-ch

	newRows := make([]interface{}, 0)
	for _, osType := range []datasource.OsType{datasource.OsTypeLinux, datasource.OsTypeDarwin} { // homebrew only provide linux and macos data
		lastCount, err := kustohelper.QueryDownloadCount(ctx, kustoClient, dbName, time.Now().AddDate(0, 0, -30), osType)
		if err != nil {
			logError(err)
		}

		requireNewBasicCount := false
		if lastCount == -1 {
			log.Printf("[HomeBrew] There is no count for %s saved on 30days ago, will generate a new basic row for it.\n", osType)
			requireNewBasicCount = true
		}

		for _, hbRaw := range hbResp {
			hb := hbRaw.(datasource.HomeBrewVersion)
			if hb.OsType != osType {
				continue
			}

			if requireNewBasicCount && hb.DataType == datasource.OneYear {
				newRows = append(newRows, datasource.HomeBrewVersion{
					OsType:        osType,
					CountDate:     time.Now(),
					DownloadCount: hb.DownloadCount,
				})
				continue
			}

			if !requireNewBasicCount && hb.DataType == datasource.ThirtyDays {
				newRows = append(newRows, datasource.HomeBrewVersion{
					OsType:        osType,
					CountDate:     time.Now(),
					DownloadCount: lastCount + hb.DownloadCount,
				})
			}
		}
	}

	if len(newRows) > 0 {
		log.Println("save homebrew basic data to kusto")
		ch, err = kustohelper.SaveToKusto(ctx, kustoClient, dbName, hbTable, kustohelper.HomeBrewIngestMapping, newRows)
		if err != nil {
			logError(fmt.Errorf("HomeBrew:%s", err))
		}
		<-ch
	}
}

func logError(err error) {
	log.Println(err)
	os.Exit(1)
}
