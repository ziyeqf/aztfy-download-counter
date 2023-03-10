package kustohelper

import (
	"context"
	"encoding/json"
	"io"
	"log"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
)

func AuthKusto(clusterUri string, clientId string, clientSecret string, tenantId string) (*kusto.Client, error) {
	kcsb := kusto.NewConnectionStringBuilder(clusterUri).WithAadAppKey(clientId, clientSecret, tenantId)
	client, err := kusto.New(kcsb)

	return client, err
}

func SaveToKusto(ctx context.Context, kustoClient *kusto.Client, dbName string, tableName string, mapping string, versions []interface{}) (chan error, error) {
	reader, encodeFunc := VersionEncode(versions)
	go encodeFunc()

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
		//ingest.FlushImmediately(), // maybe it's not a good practice.
	)
	if err != nil {
		return nil, err
	}

	return result.Wait(ctx), nil
}

func VersionEncode(data []interface{}) (*io.PipeReader, func()) {
	r, w := io.Pipe()
	enc := json.NewEncoder(w)

	return r, func() {
		defer func(w *io.PipeWriter) {
			err := w.Close()
			if err != nil {
				log.Println(err)
			}
		}(w)
		for _, d := range data {
			if err := enc.Encode(d); err != nil {
				log.Fatal(err)
			}
		}

	}
}
