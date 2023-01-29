package kustohelper

import (
	"context"
	"fmt"
	"time"

	"aztfy-download-counter/datasource"
	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/errors"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
	"github.com/Azure/azure-kusto-go/kusto/data/types"
)

func buildDefinitions() kusto.Definitions {
	defMap := map[string]kusto.ParamType{
		"countDate": {Type: types.DateTime},
		"osType":    {Type: types.String},
	}

	return kusto.NewDefinitions().Must(defMap)
}

func buildParams(date time.Time, osType datasource.OsType) kusto.Parameters {
	dateWithoutTime, _ := time.Parse(datasource.TimeFormat, date.Format(datasource.TimeFormat))
	paramMap := map[string]interface{}{
		"countDate": dateWithoutTime,
		"osType":    string(osType),
	}

	return kusto.NewParameters().Must(paramMap)
}

// it's fixed on HomeBrew table
func buildQuery(date time.Time, osType datasource.OsType) (kusto.Stmt, error) {
	stmt := kusto.NewStmt("Homebrew | where CountDate == countDate and OsType == osType")
	stmt, err := stmt.WithDefinitions(buildDefinitions())
	if err != nil {
		return stmt, err
	}

	stmt, err = stmt.WithParameters(buildParams(date, osType))
	if err != nil {
		return stmt, err
	}

	return stmt, nil
}

// QueryDownloadCount will remove the time part of date in query
// if it returns -1 without error, it means there is no existing record.
func QueryDownloadCount(ctx context.Context, kustoClient *kusto.Client, dbName string, date time.Time, osType datasource.OsType) (int32, error) {
	queryCmd, err := buildQuery(date, osType)
	if err != nil {
		return 0, err
	}

	iter, err := kustoClient.Query(ctx, dbName, queryCmd)
	if err != nil {
		return -1., err
	}
	defer iter.Stop()

	var output []datasource.HomeBrewVersion
	err = iter.DoOnRowOrError(
		func(row *table.Row, err *errors.Error) error {
			out := datasource.HomeBrewVersion{}
			if err != nil {
				return err
			}

			if err := row.ToStruct(&out); err != nil {
				return err
			}

			if row.Replace {
				output = output[:0]
			}
			output = append(output, out)
			return nil
		},
	)
	if err != nil {
		return -1, err
	}

	if len(output) == 0 {
		return -1, nil
	}

	if len(output) > 1 {
		return -1, fmt.Errorf("get more than one row, result count:%d, query stmt: %s", len(output), queryCmd.String())
	}

	return output[0].DownloadCount, nil
}
