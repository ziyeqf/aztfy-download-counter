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
	Logger    *log.Logger
	Container *azcosmos.ContainerClient
}

func (w GithubWorker) Run(ctx context.Context, date string) {
	w.Logger.Println("fetch data")
	ghResp, err := datasource.FetchGitHubDownloadCount(ctx)
	if err != nil {
		w.Logger.Println(err)
		return
	}

	w.Logger.Println("write Github data to db")
	osTypeMap := make(map[string][]database.GithubVersion)
	for _, item := range ghResp {
		item.Id = w.newGithubItemId(date, item.OsType, item.Arch, item.Ver)
		item.CountDate = date
		// getPrevObj func needs item.CountDate has value.
		prevObj, err := w.getPrevObj(ctx, w.Container, item)
		if err == nil {
			item.TodayCount = w.calcTodayCnt(prevObj, item)
		}

		array, ok := osTypeMap[item.OsType]
		if !ok {
			var array []database.GithubVersion
			osTypeMap[item.OsType] = array
		}

		osTypeMap[item.OsType] = append(array, item)
	}

	for osType, array := range osTypeMap {
		for _, item := range array {
			w.Logger.Println("update github data to db: ", item)
		}
		err = database.BatchUpsert(ctx, w.Container, osType, array)
		if err != nil {
			w.Logger.Println(err)
			return
		}
	}

	w.Logger.Println("done")
}

func (w GithubWorker) calcTodayCnt(prevObj database.GithubVersion, currObj database.GithubVersion) int {
	if prevObj.Id == "" {
		return -1
	}

	return currObj.TotalCount - prevObj.TotalCount
}

func (w GithubWorker) getPrevObj(ctx context.Context, container *azcosmos.ContainerClient, item database.GithubVersion) (database.GithubVersion, error) {
	prevDate := idx2DateStr(dateStr2Idx(item.CountDate) - 1)
	prevObj := database.GithubVersion{}
	err := database.ReadItem(ctx, container, item.OsType, w.newGithubItemId(prevDate, item.OsType, item.Arch, item.Ver), &prevObj)
	if err != nil {
		return prevObj, err
	}

	return prevObj, nil
}

func (w GithubWorker) newGithubItemId(date, osType, arch, ver string) string {
	return fmt.Sprintf("%s-%s-%s-%s", date, osType, arch, ver)
}
