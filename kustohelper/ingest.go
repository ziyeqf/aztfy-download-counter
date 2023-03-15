package kustohelper

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
)

func AuthKusto(clusterUri string, clientId string, clientSecret string, tenantId string) (*kusto.Client, error) {
	kcsb := kusto.NewConnectionStringBuilder(clusterUri).WithAadAppKey(clientId, clientSecret, tenantId)
	client, err := kusto.New(kcsb)

	return client, err
}

func SaveToKusto(ctx context.Context, kustoClient *kusto.Client, dbName string, tableName string, mapping string, versions []interface{}) (chan error, error) {
	reader, err := versionsEncode(versions)
	if err != nil {
		return nil, err
	}

	in, err := ingest.New(kustoClient, dbName, tableName)
	if err != nil {
		return nil, err
	}

	defer in.Close()

	//ingestDate := time.Now().Format("2006-01-02")
	//ingestTag := fmt.Sprintf(`["%s"]`, ingestDate)
	//save with tag will cause extremely long time pending.

	result, err := in.FromReader(ctx, reader,
		ingest.IngestionMapping(mapping, ingest.JSON),
		//ingest.IfNotExists(ingestTag),
		//ingest.Tags(append([]string{}, ingestTag)),
		ingest.ReportResultToTable(), // it's not recommended to read status, but it's helpful for debug.
		ingest.FlushImmediately(),    // maybe it's not a good practice.
	)
	if err != nil {
		return nil, err
	}

	return result.Wait(ctx), nil
}

func versionsEncode(input []interface{}) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	for _, v := range input {
		if err := enc.Encode(v); err != nil {
			return nil, err
		}
	}

	return buf, nil
}
