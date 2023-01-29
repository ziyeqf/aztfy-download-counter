package datasource

import (
	"encoding/json"
	"time"
)

type homeBrewDataType string

const (
	ThirtyDays homeBrewDataType = "30d"
	NinetyDays homeBrewDataType = "90d"
	OneYear    homeBrewDataType = "365d"
)

type HomeBrewVersion struct {
	OsType        OsType           `json:"OsType,omitempty"`
	DownloadCount int32            `json:"DownloadCount,omitempty"`
	CountDate     time.Time        `json:"CountDate,omitempty"`
	DataType      homeBrewDataType `json:"DataType,omitempty"`
}

func (h *HomeBrewVersion) MarshalJson() ([]byte, error) {
	countDateWithoutTime, err := time.Parse(TimeFormat, h.CountDate.Format(TimeFormat))
	if err != nil {
		return nil, err
	}

	return json.Marshal(&struct {
		OsType        string           `json:"OsType,omitempty"`
		DownloadCount int32            `json:"DownloadCount,omitempty"`
		CountDate     time.Time        `json:"CountDate,omitempty"`
		DataType      homeBrewDataType `json:"DataType,omitempty"`
	}{
		OsType:        string(h.OsType),
		DownloadCount: h.DownloadCount,
		CountDate:     countDateWithoutTime,
		DataType:      h.DataType,
	})
}
