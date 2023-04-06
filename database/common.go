package database

import (
	"time"
)

type OsType string

const (
	OsTypeWindows OsType = "windows"
	OsTypeLinux   OsType = "linux"
	OsTypeDarwin  OsType = "darwin" // Mac OS
)

type DBItem interface {
	HomebrewVersion | GithubVersion
}

type HomebrewVersion struct {
	Id             string `json:"id"`
	OsType         string `json:"OsType,omitempty"`
	TodayCount     int    `json:"TodayCount"`
	ThirtyDayCount int    `json:"ThirtyDayCount"`
	NinetyDayCount int    `json:"NinetyDayCount"`
	OneYearCount   int    `json:"OneYearCount"`
	ApiFailure     bool   `json:"ApiFailure"`
	CountDate      string `json:"CountDate,omitempty"`
}

type GithubVersion struct {
	Id          string    `json:"id"`
	Ver         string    `json:"Version,omitempty"`
	OsType      string    `json:"OsType,omitempty"`
	Arch        string    `json:"Arch,omitempty"`
	TodayCount  int       `json:"TodayCount,omitempty"`
	TotalCount  int       `json:"DownloadCount,omitempty"`
	PublishDate time.Time `json:"PublishDate,omitempty"`
	CountDate   string    `json:"CountDate,omitempty"`
}
