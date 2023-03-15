package kustohelper

import (
	"context"
	"time"

	"aztfy-download-counter/datasource"
	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/errors"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
)

func BuildQueryForGithubRaw(date time.Time) (kusto.Stmt, error) {
	stmt := kusto.NewStmt("GithubRaw | where CountDate >= countDate and CountDate < datetime_add('day',1, countDate)")
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

func BuildQueryForGithub(date time.Time) (kusto.Stmt, error) {
	stmt := kusto.NewStmt("Github | where CountDate >= countDate and CountDate < datetime_add('day',1, countDate)")
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

// QueryExistingRecordForGithubRaw will remove the time part of date in query
func QueryExistingRecordForGithubRaw(ctx context.Context, kustoClient *kusto.Client, dbName string, date time.Time) (data []datasource.GithubVersion, err error) {
	var output []datasource.GithubVersion

	queryCmd, err := BuildQueryForGithubRaw(date)
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
			out := datasource.GithubVersion{}
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

func QueryRunIdForGithubRaw(ctx context.Context, kustoClient *kusto.Client, dbName string) (int32, error) {
	queryCmd := kusto.NewStmt("GithubRaw | summarize max(RunId)")
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

// QueryExistingRecordForGithub will remove the time part of date in query
func QueryExistingRecordForGithub(ctx context.Context, kustoClient *kusto.Client, dbName string, date time.Time) (data []datasource.GithubVersion, err error) {
	var output []datasource.GithubVersion

	queryCmd, err := BuildQueryForGithub(date)
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
			out := datasource.GithubVersion{}
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
