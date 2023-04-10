package datasource

import (
	"context"
	"time"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/errors"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
	"github.com/Azure/azure-kusto-go/kusto/data/types"
)

const PMCDBName = "Repos"

type KustoResponse struct {
	Path string `kusto:"path"`
}

func authKusto(endpoint string) (client *kusto.Client, err error) {
	kustoConnectionStringBuilder := kusto.NewConnectionStringBuilder(endpoint)
	kustoConnectionString := kustoConnectionStringBuilder.WithAzCli()
	return kusto.New(kustoConnectionString)
}

func queryCmdAztfy(date time.Time) kusto.Stmt {
	defMap := map[string]kusto.ParamType{
		"targetDate": {Type: types.DateTime},
	}
	paramMap := map[string]interface{}{
		"targetDate": date,
	}

	return kusto.NewStmt(`HttpAccessLog 
| where path contains "aztfy"
    and path contains "rpm"
    and method == "GET"
    and code == "200"
    and (PreciseTimeStamp >= targetDate and PreciseTimeStamp <= datetime_add('day', 1, targetDate))`).MustDefinitions(kusto.NewDefinitions().Must(defMap)).MustParameters(kusto.NewParameters().Must(paramMap))
}

func queryCmdAztfexport(date time.Time) kusto.Stmt {
	defMap := map[string]kusto.ParamType{
		"targetDate": {Type: types.DateTime},
	}
	paramMap := map[string]interface{}{
		"targetDate": date,
	}

	return kusto.NewStmt(`HttpAccessLog 
| where path contains "aztfexport"
    and path contains "rpm"
    and method == "GET"
    and code == "200"
    and (PreciseTimeStamp >= targetDate and PreciseTimeStamp <= datetime_add('day', 1, targetDate))`).MustDefinitions(kusto.NewDefinitions().Must(defMap)).MustParameters(kusto.NewParameters().Must(paramMap))
}

func QueryForPMC(ctx context.Context, endpoint string, date time.Time) ([]KustoResponse, error) {
	var recs []KustoResponse

	client, err := authKusto(endpoint)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	aztfy, err := doQuery(ctx, client, queryCmdAztfy(date))
	if err != nil {
		return nil, err
	}
	recs = append(recs, aztfy...)

	aztfexport, err := doQuery(ctx, client, queryCmdAztfexport(date))
	if err != nil {
		return nil, err
	}
	recs = append(recs, aztfexport...)

	return recs, err
}

func doQuery(ctx context.Context, client *kusto.Client, cmd kusto.Stmt) ([]KustoResponse, error) {
	var recs []KustoResponse

	iter, err := client.Query(ctx, PMCDBName, cmd)
	if err != nil {
		return nil, err
	}

	err = iter.DoOnRowOrError(
		func(row *table.Row, err *errors.Error) error {
			if err != nil {
				return err
			}
			rec := KustoResponse{}
			if err := row.ToStruct(&rec); err != nil {
				return err
			}
			recs = append(recs, rec)

			return nil
		})
	return recs, nil
}
