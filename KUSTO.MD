Github:
    - create table:  .create table Github ( Version:string, OsType:string,Arch:string,DownloadCount:int, PublishDate: datetime, CountDate:datetime)
    - create mapping: .create table ['Github'] ingestion json mapping 'import_mapping' '[{"column":"Version", "Properties":{"Path":"$[\'Version\']"}},{"column":"OsType", "Properties":{"Path":"$[\'OsType\']"}},{"column":"Arch", "Properties":{"Path":"$[\'Arch\']"}},{"column":"DownloadCount", "Properties":{"Path":"$[\'DownloadCount\']"}},{"column":"PublishDate", "Properties":{"Path":"$[\'PublishDate\']"}},{"column":"CountDate", "Properties":{"Path":"$[\'CountDate\']"}}]'

HomeBrewRaw:
    - create table: .create table HomebrewRaw (DataType:string, OsType: string, DownloadCount: int, CountDate:datetime)
    - create mapping: .create table ['HomebrewRaw'] ingestion json mapping 'HomebrewRaw_mapping' '[{"column":"DataType", "Properties":{"Path":"$[\'DataType\']"}},{"column":"OsType", "Properties":{"Path":"$[\'OsType\']"}},{"column":"DownloadCount", "Properties":{"Path":"$[\'DownloadCount\']"}},{"column":"CountDate", "Properties":{"Path":"$[\'CountDate\']"}}]'

HomeBrew:
    - create table:  .create table Homebrew (OsType:string, DownloadCount:int, CountDate: datetime)
    - when there is no `30d-ago` data, it will use 365d data as a basic data.
    - create mapping: .create table ['Homebrew'] ingestion json mapping 'Homebrew_mapping' '[{"column":"OsType", "Properties":{"Path":"$[\'OsType\']"}},{"column":"DownloadCount", "Properties":{"Path":"$[\'DownloadCount\']"}},{"column":"CountDate", "Properties":{"Path":"$[\'CountDate\']"}}]'