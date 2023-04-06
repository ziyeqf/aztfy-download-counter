package datasource

import (
	"encoding/json"
	"io"
	"net/http"

	"aztfy-download-counter/database"
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
	Aztfy int `json:"aztfexport"`
}

func FetchHomeBrewDownloadCount(date string) ([]database.HomebrewVersion, error) {
	output := make([]database.HomebrewVersion, 0)

	brewJson, err := requestHomeBrewSource()
	if err != nil {
		return output, err
	}

	for _, osType := range []database.OsType{database.OsTypeDarwin, database.OsTypeLinux} {
		output = append(output, generateHomeBrewVersion(*brewJson, osType, date))
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

func generateHomeBrewVersion(input BrewJson, osType database.OsType, date string) database.HomebrewVersion {
	var i install
	if osType == database.OsTypeDarwin {
		i = input.Analytics.Install
	} else {
		i = input.AnalyticsLinux.Install
	}

	output := database.HomebrewVersion{
		OsType:         string(osType),
		CountDate:      date,
		ThirtyDayCount: i.ThirtyDays.Aztfy,
		NinetyDayCount: i.NinetyDays.Aztfy,
		OneYearCount:   i.OneYear.Aztfy,
	}

	return output
}
