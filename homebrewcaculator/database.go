package homebrewcaculator

const RecordStartDate = "2023-03-24" // data before this date should be filled with zero value

type DatabaseClient interface {
	IsAValidDate(date Date) bool                 //if the date is not valid?
	GetDataAtDayN(date Date) (Data, error)       // do not return error when there is no record.
	UpdateDataAtDayN(date Date, data Data) error // add a new record or update the existing record.
}
