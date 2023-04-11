package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

func AuthDBClient(endpoint string, key string, dbName string) (*azcosmos.DatabaseClient, error) {
	cred, err := azcosmos.NewKeyCredential(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create a credential: %+v", err)
	}

	client, err := azcosmos.NewClientWithKey(endpoint, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Cosmos DB client: %+v", err)
	}

	dbClient, err := client.NewDatabase(dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to create a database client: %+v", err)
	}

	return dbClient, nil
}

func CreateOrUpdateItem[T DBItem](ctx context.Context, container *azcosmos.ContainerClient, osType string, item T) error {
	pk := azcosmos.NewPartitionKeyString(osType)

	b, err := json.Marshal(item)
	if err != nil {
		return err
	}

	itemOptions := azcosmos.ItemOptions{
		ConsistencyLevel: azcosmos.ConsistencyLevelSession.ToPtr(),
	}

	_, err = container.UpsertItem(ctx, pk, b, &itemOptions)

	if err != nil {
		return err
	}

	return nil
}

func QueryItem[T DBItem](ctx context.Context, container *azcosmos.ContainerClient, osType, date string, response T) (resp []T, err error) {
	pk := azcosmos.NewPartitionKeyString(osType)

	opt := azcosmos.QueryOptions{
		QueryParameters: []azcosmos.QueryParameter{
			{"@ostype", osType},
			{"@cntDate", date},
		},
	}

	queryPager := container.NewQueryItemsPager("select * from c where c.OsType = @ostype AND c.Date = @cntDate", pk, &opt)

	var errs error
	for queryPager.More() {
		queryResponse, err := queryPager.NextPage(ctx)
		if err != nil {
			errs = errors.Join(err)
		}

		for _, item := range queryResponse.Items {
			err := json.Unmarshal(item, &response)
			if err != nil {
				errs = errors.Join(err)
				continue
			}
			resp = append(resp, response)
		}
	}
	return resp, errs
}

func BatchUpsert[T DBItem](ctx context.Context, container *azcosmos.ContainerClient, pkStr string, items []T) error {
	pk := azcosmos.NewPartitionKeyString(pkStr)

	batch := container.NewTransactionalBatch(pk)

	for _, item := range items {
		b, err := json.Marshal(item)
		if err != nil {
			return err
		}

		batch.UpsertItem(b, nil)
	}

	response, err := container.ExecuteTransactionalBatch(ctx, batch, nil)
	if err != nil {
		return err
	}

	if !response.Success {
		var err error
		for _, item := range response.OperationResults {
			err = errors.Join(err, fmt.Errorf("insert failed, code: %v", item.StatusCode))
		}
		return err
	}

	return nil
}

func ReadItem[T DBItem](ctx context.Context, container *azcosmos.ContainerClient, pkStr, itemId string, response *T) error {
	pk := azcosmos.NewPartitionKeyString(pkStr)

	itemResponse, err := container.ReadItem(ctx, pk, itemId, nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(itemResponse.Value, &response)
	if err != nil {
		return err
	}

	return nil
}
