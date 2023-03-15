package datasource

import (
	"encoding/json"
	"time"
)

type HomeBrewDataType string

const (
	ThirtyDays HomeBrewDataType = "30d"
	NinetyDays HomeBrewDataType = "90d"
	OneYear    HomeBrewDataType = "365d"
)

type HomeBrewVersion struct {
	RunId         int32     `json:"RunId,omitempty"`
	OsType        string    `json:"OsType,omitempty"`
	DownloadCount int32     `json:"DownloadCount"`
	CountDate     time.Time `json:"CountDate,omitempty"`
	DataType      string    `json:"DataType,omitempty"`
}

func (h *HomeBrewVersion) MarshalJson() ([]byte, error) {
	countDateWithoutTime, err := time.Parse(TimeFormat, h.CountDate.Format(TimeFormat))
	if err != nil {
		return nil, err
	}

	return json.Marshal(&struct {
		RunId         int32     `json:"RunId,omitempty"`
		OsType        string    `json:"OsType,omitempty"`
		DownloadCount int32     `json:"DownloadCount,omitempty"`
		CountDate     time.Time `json:"CountDate,omitempty"`
		DataType      string    `json:"DataType,omitempty"`
	}{
		RunId:         h.RunId,
		OsType:        h.OsType,
		DownloadCount: h.DownloadCount,
		CountDate:     countDateWithoutTime,
		DataType:      h.DataType,
	})
}
