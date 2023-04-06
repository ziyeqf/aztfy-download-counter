package homebrewcaculator

import "context"

// CountInfo represents the count information of a certain index that is recorded by the DatabasesClient.
type CountInfo struct {
	// Count represents the count at this index. Especially, -1 represents there is no count at this index.
	Count int
	// TotalCounts represents the cumulative counts of each span.
	TotalCounts map[Span]int
}

type DatabaseClient interface {
	Get(ctx context.Context, idx int) (CountInfo, error)    // do not return error when there is no record.
	Set(ctx context.Context, idx int, data CountInfo) error // add a new record or update the existing record.
}
