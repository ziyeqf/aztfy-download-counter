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

type TotalCountResponse struct {
	Count int64 `kusto:"Count"`
}

func AuthKusto(clientId, clientSecret, tenantId, endpoint string) (client *kusto.Client, err error) {
	kustoConnectionStringBuilder := kusto.NewConnectionStringBuilder(endpoint)
	kustoConnectionString := kustoConnectionStringBuilder.WithAadAppKey(clientId, clientSecret, tenantId)
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
    and (PreciseTimeStamp >= datetime_add('day',-1,targetDate) and PreciseTimeStamp <= targetDate)`).MustDefinitions(kusto.NewDefinitions().Must(defMap)).MustParameters(kusto.NewParameters().Must(paramMap))
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
    and (PreciseTimeStamp >= datetime_add('day',-1,targetDate) and PreciseTimeStamp <= targetDate)`).MustDefinitions(kusto.NewDefinitions().Must(defMap)).MustParameters(kusto.NewParameters().Must(paramMap))
}

func queryCmdForTotalCountAztfexport(startDate, endDate time.Time, arch, version string) kusto.Stmt {
	defMap := map[string]kusto.ParamType{
		"startDate":     {Type: types.DateTime},
		"endDate":       {Type: types.DateTime},
		"targetArch":    {Type: types.String},
		"targetVersion": {Type: types.String},
	}
	paramMap := map[string]interface{}{
		"startDate":     startDate,
		"endDate":       endDate,
		"targetArch":    arch,
		"targetVersion": version,
	}

	return kusto.NewStmt(`HttpAccessLog 
| where path contains "aztfexport"
	and path contains targetArch
	and path contains targetVersion
    and path contains "rpm"
    and method == "GET"
    and code == "200"
    and (PreciseTimeStamp >= startDate and PreciseTimeStamp <= endDate)
| count `).MustDefinitions(kusto.NewDefinitions().Must(defMap)).MustParameters(kusto.NewParameters().Must(paramMap))
}

func queryCmdForTotalCountAztfy(startDate, endDate time.Time, arch, version string) kusto.Stmt {
	defMap := map[string]kusto.ParamType{
		"startDate":     {Type: types.DateTime},
		"endDate":       {Type: types.DateTime},
		"targetArch":    {Type: types.String},
		"targetVersion": {Type: types.String},
	}
	paramMap := map[string]interface{}{
		"startDate":     startDate,
		"endDate":       endDate,
		"targetArch":    arch,
		"targetVersion": version,
	}

	return kusto.NewStmt(`HttpAccessLog 
| where path contains "aztfy"
	and path contains targetArch
	and path contains targetVersion
    and path contains "rpm"
    and method == "GET"
    and code == "200"
    and (PreciseTimeStamp >= startDate and PreciseTimeStamp <= endDate)
| count `).MustDefinitions(kusto.NewDefinitions().Must(defMap)).MustParameters(kusto.NewParameters().Must(paramMap))
}

func QueryTotalCount(ctx context.Context, client *kusto.Client, startDate, endDate time.Time, version, arch string) (int64, error) {
	aztfy, err := doCntQuery(ctx, client, queryCmdForTotalCountAztfexport(startDate, endDate, arch, version))
	if err != nil {
		return -1, err
	}

	aztfexport, err := doCntQuery(ctx, client, queryCmdForTotalCountAztfy(startDate, endDate, arch, version))
	if err != nil {
		return -1, err
	}

	var ret int64 = 0
	if len(aztfy) > 0 {
		ret += aztfy[0].Count
	}
	if len(aztfexport) > 0 {
		ret += aztfexport[0].Count
	}

	return ret, nil
}

func doCntQuery(ctx context.Context, client *kusto.Client, cmd kusto.Stmt) ([]TotalCountResponse, error) {
	var recs []TotalCountResponse

	iter, err := client.Query(ctx, PMCDBName, cmd)
	if err != nil {
		return nil, err
	}

	err = iter.DoOnRowOrError(
		func(row *table.Row, err *errors.Error) error {
			if err != nil {
				return err
			}
			rec := TotalCountResponse{}
			if err := row.ToStruct(&rec); err != nil {
				return err
			}
			recs = append(recs, rec)

			return nil
		},
	)

	return recs, err
}

func QueryForPMC(ctx context.Context, client *kusto.Client, date time.Time) ([]KustoResponse, error) {
	var recs []KustoResponse

	aztfy, err := doPMCQuery(ctx, client, queryCmdAztfy(date))
	if err != nil {
		return nil, err
	}
	recs = append(recs, aztfy...)

	aztfexport, err := doPMCQuery(ctx, client, queryCmdAztfexport(date))
	if err != nil {
		return nil, err
	}
	recs = append(recs, aztfexport...)

	return recs, err
}

func doPMCQuery(ctx context.Context, client *kusto.Client, cmd kusto.Stmt) ([]KustoResponse, error) {
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
