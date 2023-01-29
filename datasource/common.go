package datasource

const TimeFormat = "2006-01-02"

type OsType string

const (
	OsTypeWindows OsType = "windows"
	OsTypeLinux   OsType = "linux"
	OsTypeDarwin  OsType = "darwin"  // Mac OS
	OsTypeFreeBSD OsType = "freebsd" //deprecated
)
