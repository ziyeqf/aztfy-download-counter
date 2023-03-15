package datasource

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const HomeBrewApiUri = "https://formulae.brew.sh/api/formula/aztfexport.json"

type BrewJson struct {
	Analytics struct {
		Install install `json:"install"`
	} `json:"analytics"`

	AnalyticsLinux struct {
		Install install `json:"install"`
	} `json:"analytics-linux"`
}

type install struct {
	ThirtyDays installCount `json:"30d"`
	NinetyDays installCount `json:"90d"`
	OneYear    installCount `json:"365d"`
}

type installCount struct {
	Aztfy int `json:"aztfy"`
}

func FetchHomeBrewDownloadCount(runId int32) ([]interface{}, error) {
	output := make([]interface{}, 0)

	brewJson, err := requestHomeBrewSource()
	if err != nil {
		return output, err
	}

	for _, osType := range []OsType{OsTypeDarwin, OsTypeLinux} {
		for _, dataType := range []HomeBrewDataType{ThirtyDays, NinetyDays, OneYear} {
			output = append(output, generateHomeBrewVersion(*brewJson, osType, dataType, runId))
		}
	}

	return output, nil
}

func requestHomeBrewSource() (*BrewJson, error) {
	cli := http.Client{}
	resp, err := cli.Get(HomeBrewApiUri)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var brewJson BrewJson

	decErr := json.NewDecoder(resp.Body).Decode(&brewJson)
	if decErr == io.EOF {
		decErr = nil
	}
	if decErr != nil {
		err = decErr
	}

	return &brewJson, err
}

func generateHomeBrewVersion(input BrewJson, osType OsType, dataType HomeBrewDataType, runId int32) HomeBrewVersion {
	var i install
	if osType == OsTypeDarwin {
		i = input.Analytics.Install
	} else {
		i = input.AnalyticsLinux.Install
	}

	output := HomeBrewVersion{
		RunId:     runId,
		OsType:    string(osType),
		DataType:  string(dataType),
		CountDate: time.Now(),
	}

	switch dataType {
	case ThirtyDays:
		output.DownloadCount = int32(i.ThirtyDays.Aztfy)
	case NinetyDays:
		output.DownloadCount = int32(i.NinetyDays.Aztfy)
	case OneYear:
		output.DownloadCount = int32(i.OneYear.Aztfy)
	}

	return output
}
