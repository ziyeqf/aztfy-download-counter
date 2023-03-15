package datasource

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v50/github"
)

const GithubPerPage = 20

func FetchGitHubDownloadCount(ctx context.Context, runId int32) ([]interface{}, error) {
	output := make([]interface{}, 0)
	releases, err := FetchReleaseList(ctx)
	if err != nil {
		return output, err
	}

	for _, r := range releases {
		for _, a := range r.Assets {
			if a.Name == nil || a.DownloadCount == nil {
				continue
			}

			version, osType, arch, err := parseTagName(*a.Name, *a.ContentType)
			if err != nil {
				log.Printf("[github] skip assest: %s\n", *a.Name)
				continue
			}

			output = append(output, GithubVersion{
				RunId:         runId,
				Ver:           version,
				OsType:        string(osType),
				Arch:          arch,
				DownloadCount: int32(*a.DownloadCount),
				PublishDate:   r.PublishedAt.Time,
				CountDate:     time.Now(),
			})
		}
	}
	return output, nil
}

func FetchReleaseList(ctx context.Context) ([]*github.RepositoryRelease, error) {
	client := github.NewClient(nil)
	result := make([]*github.RepositoryRelease, 0)

	itemCount := 0
	for i := 1; itemCount < GithubPerPage; i++ {
		itemCount = 0

		opt := &github.ListOptions{
			Page:    i,
			PerPage: GithubPerPage,
		}

		releases, _, err := client.Repositories.ListReleases(ctx, "Azure", "aztfy", opt)
		if err != nil {
			return nil, err
		}

		result = append(result, releases...)
		itemCount = len(releases)
	}

	return result, nil
}

func parseTagName(tagName string, contentType string) (version string, osType OsType, arch string, err error) {
	switch contentType {
	case "application/zip":
		return parseTagNameForZip(tagName)
	case "application/x-msdownload":
		return parseTagNameForMsi(tagName)
	case "application/gzip":
		return parseTagNameForGz(tagName)
	default:
		return "", "", "", fmt.Errorf("parse failed")
	}
}

func parseTagNameForZip(tagName string) (version string, osType OsType, arch string, err error) {
	reg := regexp.MustCompile(`(?m).*_(v\d*\.\d*\.\d*)_([a-z]+)_(.+)\.zip`)
	result := reg.FindStringSubmatch(tagName)
	if len(result) != 4 {
		return "", "", "", fmt.Errorf("parse failed")
	}
	return result[1], OsType(strings.ToLower(result[2])), result[3], nil
}

func parseTagNameForMsi(tagName string) (version string, osType OsType, arch string, err error) {
	reg := regexp.MustCompile(`.*_(v\d*\.\d*\.\d*)_(.+)\.msi`)
	result := reg.FindStringSubmatch(tagName)
	if len(result) != 3 {
		return "", "", "", fmt.Errorf("parse failed")
	}
	return result[1], OsTypeWindows, result[2], nil
}

func parseTagNameForGz(tagName string) (version string, osType OsType, arch string, err error) {
	reg := regexp.MustCompile(`(?mU).*_(v{0,1}\d*\.\d*\.\d*)_(.+)_(.+)\.tar\.gz`)
	result := reg.FindStringSubmatch(tagName)
	if len(result) != 4 {
		return "", "", "", fmt.Errorf("parse failed")
	}
	return result[1], OsType(strings.ToLower(result[2])), result[3], nil
}
