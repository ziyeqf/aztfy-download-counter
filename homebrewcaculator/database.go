package homebrewcaculator

// CountInfo represents the count information of a certain index that is recorded by the DatabasesClient.
type CountInfo struct {
	// Count represents the count at this index. Especially, -1 represents there is no count at this index.
	Count int32
	// TotalCounts represents the cumulative counts of each span.
	TotalCounts map[Span]int32
}

type DatabaseClient interface {
	Get(idx int) (CountInfo, error)    // do not return error when there is no record.
	Set(idx int, data CountInfo) error // add a new record or update the existing record.
}
