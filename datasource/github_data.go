package datasource

import (
	"encoding/json"
	"time"
)

type GithubVersion struct {
	RunId         int32     `json:"RunId,omitempty"`
	Ver           string    `json:"Version,omitempty"`
	OsType        string    `json:"OsType,omitempty"`
	Arch          string    `json:"Arch,omitempty"`
	DownloadCount int32     `json:"DownloadCount,omitempty"`
	PublishDate   time.Time `json:"PublishDate,omitempty"`
	CountDate     time.Time `json:"CountDate,omitempty"`
}

func (h *GithubVersion) MarshalJson() ([]byte, error) {
	countDateWithoutTime, err := time.Parse(TimeFormat, h.CountDate.Format(TimeFormat))
	if err != nil {
		return nil, err
	}

	return json.Marshal(&struct {
		RunId         int32     `json:"RunId,omitempty"`
		Ver           string    `json:"Version,omitempty"`
		OsType        string    `json:"OsType,omitempty"`
		Arch          string    `json:"Arch,omitempty"`
		DownloadCount int32     `json:"DownloadCount,omitempty"`
		PublishDate   time.Time `json:"PublishDate,omitempty"`
		CountDate     time.Time `json:"CountDate,omitempty"`
	}{
		RunId:         h.RunId,
		Ver:           h.Ver,
		OsType:        h.OsType,
		Arch:          h.Arch,
		DownloadCount: h.DownloadCount,
		PublishDate:   h.PublishDate,
		CountDate:     countDateWithoutTime,
	})
}
