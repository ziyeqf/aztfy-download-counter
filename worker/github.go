package worker

import (
	"context"
	"fmt"
	"log"

	"aztfy-download-counter/database"
	"aztfy-download-counter/datasource"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

type GithubWorker struct {
	Container *azcosmos.ContainerClient
}

func (w GithubWorker) Run(ctx context.Context, date string) error {
	log.Println("[Github]fetch data")
	ghResp, err := datasource.FetchGitHubDownloadCount(ctx)
	if err != nil {
		return err
	}

	log.Println("[Github] write Github data to db")
	osTypeMap := make(map[string][]database.GithubVersion)
	for _, item := range ghResp {
		item.Id = newGithubItemId(date, item.OsType, item.Arch, item.Ver)
		prevObj, err := getPrevObj(ctx, w.Container, item.OsType, date)
		if err == nil {
			item.TodayCount = calcTodayCnt(prevObj, item)
		}

		array, ok := osTypeMap[item.OsType]
		if !ok {
			var array []database.GithubVersion
			osTypeMap[item.OsType] = array
		}

		osTypeMap[item.OsType] = append(array, item)
	}

	for osType, array := range osTypeMap {
		err = database.BatchUpsert(ctx, w.Container, osType, array)
		if err != nil {
			return err
		}
	}

	return nil
}

func calcTodayCnt(prevObj database.GithubVersion, currObj database.GithubVersion) int {
	if prevObj.Id == "" {
		return -1
	}

	return currObj.TotalCount - prevObj.TotalCount
}

func getPrevObj(ctx context.Context, container *azcosmos.ContainerClient, osType, date string) (database.GithubVersion, error) {
	var prevObj database.GithubVersion
	resp, err := database.QueryItem(ctx, container, osType, date, database.GithubVersion{})
	if err != nil {
		return prevObj, err
	}

	if len(resp) == 0 {
		return prevObj, nil
	}

	if len(resp) > 1 {
		return prevObj, fmt.Errorf("more than one item found for %s", date)
	}

	return resp[0], nil
}

func newGithubItemId(date, osType, arch, ver string) string {
	return fmt.Sprintf("%s-%s-%s-%s", date, osType, arch, ver)
}
