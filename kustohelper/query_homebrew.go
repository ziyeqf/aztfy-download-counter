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

func buildDefinitionsForHomeBrew() kusto.Definitions {
	defMap := map[string]kusto.ParamType{
		"countDate": {Type: types.DateTime},
		"osType":    {Type: types.String},
	}

	return kusto.NewDefinitions().Must(defMap)
}

func buildParamsForHomeBrew(date time.Time, osType datasource.OsType) kusto.Parameters {
	dateWithoutTime, _ := time.Parse(datasource.TimeFormat, date.Format(datasource.TimeFormat))
	paramMap := map[string]interface{}{
		"countDate": dateWithoutTime,
		"osType":    string(osType),
	}

	return kusto.NewParameters().Must(paramMap)
}

func homebrewBuildQuery(date time.Time, osType datasource.OsType) (kusto.Stmt, error) {
	stmt := kusto.NewStmt("Homebrew | where CountDate >= countDate and CountDate < datetime_add('day',1, countDate) and OsType == osType")
	stmt, err := stmt.WithDefinitions(buildDefinitionsForHomeBrew())
	if err != nil {
		return stmt, err
	}

	stmt, err = stmt.WithParameters(buildParamsForHomeBrew(date, osType))
	if err != nil {
		return stmt, err
	}

	return stmt, nil
}

func BuildExistingQueryForHomebrewRaw(date time.Time) (kusto.Stmt, error) {
	stmt := kusto.NewStmt("HomebrewRaw | where CountDate >= countDate and CountDate < datetime_add('day',1, countDate)")
	stmt, err := stmt.WithDefinitions(buildDefinitionsForExistingCheck())
	if err != nil {
		return stmt, err
	}

	stmt, err = stmt.WithParameters(buildParamsForExistingCheck(date))
	if err != nil {
		return stmt, err
	}

	return stmt, nil
}

func BuildExistingQueryForHomebrew(date time.Time) (kusto.Stmt, error) {
	stmt := kusto.NewStmt("Homebrew | where CountDate >= countDate and CountDate < datetime_add('day',1, countDate)")
	stmt, err := stmt.WithDefinitions(buildDefinitionsForExistingCheck())
	if err != nil {
		return stmt, err
	}

	stmt, err = stmt.WithParameters(buildParamsForExistingCheck(date))
	if err != nil {
		return stmt, err
	}

	return stmt, nil
}

// QueryDownloadCountForHomebrew will remove the time part of date in query
// if it returns -1 without error, it means there is no existing record.
func QueryDownloadCountForHomebrew(ctx context.Context, kustoClient *kusto.Client, dbName string, date time.Time, osType datasource.OsType) (int32, error) {
	queryCmd, err := homebrewBuildQuery(date, osType)
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

func QueryExistingRecordForHomebrew(ctx context.Context, kustoClient *kusto.Client, dbName string, date time.Time) (data []datasource.HomeBrewVersion, err error) {
	var output []datasource.HomeBrewVersion

	queryCmd, err := BuildExistingQueryForHomebrew(date)
	if err != nil {
		return output, err
	}

	iter, err := kustoClient.Query(ctx, dbName, queryCmd)
	if err != nil {
		return output, err
	}
	defer iter.Stop()

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

	return output, nil
}

func QueryExistingRecordForHomebrewRaw(ctx context.Context, kustoClient *kusto.Client, dbName string, date time.Time) (data []datasource.HomeBrewVersion, err error) {
	var output []datasource.HomeBrewVersion

	queryCmd, err := BuildExistingQueryForHomebrewRaw(date)
	if err != nil {
		return output, err
	}

	iter, err := kustoClient.Query(ctx, dbName, queryCmd)
	if err != nil {
		return output, err
	}
	defer iter.Stop()

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

	return output, nil
}

func QueryRunIdForHomeBrewRaw(ctx context.Context, kustoClient *kusto.Client, dbName string) (int32, error) {
	queryCmd := kusto.NewStmt("HomebrewRaw | summarize max(RunId)")
	iter, err := kustoClient.Query(ctx, dbName, queryCmd)
	if err != nil {
		return 0, err
	}

	runId := RunId{
		RunId: 0,
	}
	err = iter.DoOnRowOrError(
		func(row *table.Row, err *errors.Error) error {
			if err != nil {
				return err
			}

			if err := row.ToStruct(&runId); err != nil {
				return err
			}

			return nil
		})

	return runId.RunId, nil
}

func QueryRunIdForHomeBrew(ctx context.Context, kustoClient *kusto.Client, dbName string) (int32, error) {
	queryCmd := kusto.NewStmt("Homebrew | summarize max(RunId)")
	iter, err := kustoClient.Query(ctx, dbName, queryCmd)
	if err != nil {
		return 0, err
	}

	runId := RunId{
		RunId: 0,
	}
	err = iter.DoOnRowOrError(
		func(row *table.Row, err *errors.Error) error {
			if err != nil {
				return err
			}

			if err := row.ToStruct(&runId); err != nil {
				return err
			}

			return nil
		})

	return runId.RunId, nil
}
