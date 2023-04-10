package datasource

import (
	"encoding/json"
	"io"
	"net/http"
)

const HomeBrewApiUri = "https://formulae.brew.sh/api/formula/aztfexport.json"

type BrewJson struct {
	Analytics struct {
		Install Install `json:"Install"`
	} `json:"analytics"`

	AnalyticsLinux struct {
		Install Install `json:"Install"`
	} `json:"analytics-linux"`
}

type Install struct {
	ThirtyDays InstallCount `json:"30d"`
	NinetyDays InstallCount `json:"90d"`
	OneYear    InstallCount `json:"365d"`
}

type InstallCount struct {
	Aztfy int `json:"aztfexport"`
}

func FetchHomeBrewDownloadCount() (*BrewJson, error) {
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
