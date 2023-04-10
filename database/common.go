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
	HomebrewVersion | GithubVersion | PMCVersion
}

type HomebrewVersion struct {
	Id             string `json:"id"`
	OsType         string `json:"OsType"`
	TodayCount     int    `json:"TodayCount"`
	ThirtyDayCount int    `json:"ThirtyDayCount"`
	NinetyDayCount int    `json:"NinetyDayCount"`
	OneYearCount   int    `json:"OneYearCount"`
	ApiFailure     bool   `json:"ApiFailure"`
	CountDate      string `json:"Date"`
}

type GithubVersion struct {
	Id          string    `json:"id"`
	Ver         string    `json:"Version"`
	OsType      string    `json:"OsType"`
	Arch        string    `json:"Arch"`
	TodayCount  int       `json:"TodayCount"`
	TotalCount  int       `json:"DownloadCount"`
	PublishDate time.Time `json:"PublishDate"`
	CountDate   string    `json:"Date"`
}

type PMCVersion struct {
	Id         string `json:"id"`
	Ver        string `json:"Version"`
	Arch       string `json:"Arch"`
	TodayCount int    `json:"TodayCount"`
	Date       string `json:"Date"`
}
