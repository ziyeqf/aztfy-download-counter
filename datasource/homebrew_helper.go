package datasource

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const HomeBrewApiUri = "https://formulae.brew.sh/api/formula/aztfy.json"

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

func FetchHomeBrewDownloadCount() ([]HomeBrewVersion, error) {
	output := make([]HomeBrewVersion, 0)

	brewJson, err := requestHomeBrewSource()
	if err != nil {
		return output, err
	}

	for _, osType := range []OsType{OsTypeDarwin, OsTypeLinux} {
		for _, dataType := range []homeBrewDataType{ThirtyDays, NinetyDays, OneYear} {
			output = append(output, generateHomeBrewVersion(brewJson.Analytics.Install, osType, dataType))
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

func generateHomeBrewVersion(input install, osType OsType, dataType homeBrewDataType) HomeBrewVersion {
	output := HomeBrewVersion{
		OsType:    osType,
		DataType:  dataType,
		CountDate: time.Now(),
	}

	switch dataType {
	case ThirtyDays:
		output.DownloadCount = int32(input.ThirtyDays.Aztfy)
	case NinetyDays:
		output.DownloadCount = int32(input.NinetyDays.Aztfy)
	case OneYear:
		output.DownloadCount = int32(input.OneYear.Aztfy)
	}

	return output
}
