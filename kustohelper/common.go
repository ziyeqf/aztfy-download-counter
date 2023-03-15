package kustohelper

import (
	"time"

	"aztfy-download-counter/datasource"
	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/types"
)

type RunId struct {
	RunId int32 `json:"max_RunId"`
}

func buildDefinitionsForExistingCheck() kusto.Definitions {
	defMap := map[string]kusto.ParamType{
		"countDate": {Type: types.DateTime},
	}

	return kusto.NewDefinitions().Must(defMap)
}

func buildParamsForExistingCheck(date time.Time) kusto.Parameters {
	dateWithoutTime, _ := time.Parse(datasource.TimeFormat, date.Format(datasource.TimeFormat))
	paramMap := map[string]interface{}{
		"countDate": dateWithoutTime,
	}

	return kusto.NewParameters().Must(paramMap)
}
