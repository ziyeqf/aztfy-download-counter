package job

import (
	"context"
	"fmt"
	"log"

	"aztfy-download-counter/database"
	"aztfy-download-counter/datasource"
	"aztfy-download-counter/job/githubutils"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/google/go-github/v50/github"
)

type GithubWorker struct {
	ContainerInitFunc func() (*azcosmos.ContainerClient, error)
	Logger            *log.Logger
	Date              string
}

func (w GithubWorker) Run(ctx context.Context) {
	container, err := w.ContainerInitFunc()
	if err != nil {
		w.Logger.Println(err)
		return
	}

	w.Logger.Println("fetch data")
	ghResp, err := datasource.FetchGitHubDownloadCount(ctx)
	if err != nil {
		w.Logger.Println(err)
		return
	}
	items := w.processReleases(ghResp, w.Date)

	w.Logger.Println("write Github data to db")
	osTypeMap := make(map[string][]database.GithubVersion)
	for _, item := range items {
		prevObj, err := w.getPrevObj(ctx, container, item)
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
		err = database.BatchUpsert(ctx, container, osType, array)
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

func (w GithubWorker) processReleases(releases []*github.RepositoryRelease, countDate string) []database.GithubVersion {
	var output []database.GithubVersion
	for _, r := range releases {
		for _, a := range r.Assets {
			if a.Name == nil || a.DownloadCount == nil {
				continue
			}

			version, osType, arch, err := githubutils.ParseTagName(*a.Name, *a.ContentType)
			if err != nil {
				continue
			}

			output = append(output, database.GithubVersion{
				Id:          w.newGithubItemId(countDate, string(osType), arch, version),
				CountDate:   countDate,
				Ver:         version,
				OsType:      string(osType),
				Arch:        arch,
				TotalCount:  *a.DownloadCount,
				PublishDate: r.PublishedAt.Time,
			})
		}
	}
	return output
}
