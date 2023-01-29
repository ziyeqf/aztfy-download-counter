package datasource

import (
	"encoding/json"
	"time"
)

type GithubVersion struct {
	Ver           string    `json:"Version,omitempty"`
	OsType        OsType    `json:"OsType,omitempty"`
	Arch          string    `json:"Arch,omitempty"`
	DownloadCount int32     `json:"DownloadCount,omitempty"`
	PublishDate   string    `json:"PublishDate,omitempty"`
	CountDate     time.Time `json:"CountDate,omitempty"`
}

func (h *GithubVersion) MarshalJson() ([]byte, error) {
	countDateWithoutTime, err := time.Parse(TimeFormat, h.CountDate.Format(TimeFormat))
	if err != nil {
		return nil, err
	}

	return json.Marshal(&struct {
		Ver           string    `json:"Version,omitempty"`
		OsType        string    `json:"OsType,omitempty"`
		Arch          string    `json:"Arch,omitempty"`
		DownloadCount int32     `json:"DownloadCount,omitempty"`
		PublishDate   string    `json:"PublishDate,omitempty"`
		CountDate     time.Time `json:"CountDate,omitempty"`
	}{
		Ver:           h.Ver,
		OsType:        string(h.OsType),
		Arch:          h.Arch,
		DownloadCount: h.DownloadCount,
		PublishDate:   h.PublishDate,
		CountDate:     countDateWithoutTime,
	})
}
