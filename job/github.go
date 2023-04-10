package job

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"aztfy-download-counter/database"
	"aztfy-download-counter/datasource"
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
		for _, item := range array {
			w.Logger.Println("update github data to db: ", item)
		}
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

			version, osType, arch, err := w.parseTagName(*a.Name, *a.ContentType)
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

func (w GithubWorker) parseTagName(tagName string, contentType string) (version string, osType database.OsType, arch string, err error) {
	switch contentType {
	case "application/zip":
		return w.parseTagNameForZip(tagName)
	case "application/x-msdownload":
		return w.parseTagNameForMsi(tagName)
	case "application/gzip":
		return w.parseTagNameForGz(tagName)
	default:
		return "", "", "", fmt.Errorf("parse failed")
	}
}

func (w GithubWorker) parseTagNameForZip(tagName string) (version string, osType database.OsType, arch string, err error) {
	reg := regexp.MustCompile(`(?m).*_(v\d*\.\d*\.\d*)_([a-z]+)_(.+)\.zip`)
	result := reg.FindStringSubmatch(tagName)
	if len(result) != 4 {
		return "", "", "", fmt.Errorf("parse failed")
	}
	return result[1], database.OsType(strings.ToLower(result[2])), result[3], nil
}

func (w GithubWorker) parseTagNameForMsi(tagName string) (version string, osType database.OsType, arch string, err error) {
	reg := regexp.MustCompile(`.*_(v\d*\.\d*\.\d*)_(.+)\.msi`)
	result := reg.FindStringSubmatch(tagName)
	if len(result) != 3 {
		return "", "", "", fmt.Errorf("parse failed")
	}
	return result[1], database.OsTypeWindows, result[2], nil
}

func (w GithubWorker) parseTagNameForGz(tagName string) (version string, osType database.OsType, arch string, err error) {
	reg := regexp.MustCompile(`(?mU).*_(v{0,1}\d*\.\d*\.\d*)_(.+)_(.+)\.tar\.gz`)
	result := reg.FindStringSubmatch(tagName)
	if len(result) != 4 {
		return "", "", "", fmt.Errorf("parse failed")
	}
	return result[1], database.OsType(strings.ToLower(result[2])), result[3], nil
}
