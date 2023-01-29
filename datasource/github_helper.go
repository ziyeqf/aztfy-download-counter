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

func FetchGitHubDownloadCount(ctx context.Context) ([]interface{}, error) {
	output := make([]interface{}, 0)
	releases, err := FetchReleaseList(ctx, 1, 20)
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
				Ver:           version,
				OsType:        osType,
				Arch:          arch,
				DownloadCount: int32(*a.DownloadCount),
				PublishDate:   r.PublishedAt.Format(TimeFormat),
				CountDate:     time.Now(),
			})
		}
	}
	return output, nil
}

func FetchReleaseList(ctx context.Context, page int, perPage int) ([]*github.RepositoryRelease, error) {
	client := github.NewClient(nil)

	opt := &github.ListOptions{
		Page:    page,
		PerPage: perPage,
	}

	releases, _, err := client.Repositories.ListReleases(ctx, "Azure", "aztfy", opt)
	if err != nil {
		return nil, err
	}

	return releases, nil
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
	reg := regexp.MustCompile(`(?m)aztfy_(v\d\.\d\.\d)_([a-z]+)_(.+)\.zip`)
	result := reg.FindStringSubmatch(tagName)
	if len(result) != 4 {
		return "", "", "", fmt.Errorf("parse failed")
	}
	return result[1], OsType(strings.ToLower(result[2])), result[3], nil
}

func parseTagNameForMsi(tagName string) (version string, osType OsType, arch string, err error) {
	reg := regexp.MustCompile(`aztfy_(v\d\.\d\.\d)_(.+)\.msi`)
	result := reg.FindStringSubmatch(tagName)
	if len(result) != 3 {
		return "", "", "", fmt.Errorf("parse failed")
	}
	return result[1], OsTypeWindows, result[2], nil
}

func parseTagNameForGz(tagName string) (version string, osType OsType, arch string, err error) {
	reg := regexp.MustCompile(`(?mU)aztfy_(v{0,1}\d\.\d\.\d)_(.+)\_(.+)\.tar\.gz`)
	result := reg.FindStringSubmatch(tagName)
	if len(result) != 4 {
		return "", "", "", fmt.Errorf("parse failed")
	}
	return result[1], OsType(strings.ToLower(result[2])), result[3], nil
}
