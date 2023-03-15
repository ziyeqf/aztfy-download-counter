GithubRaw:
    - create table:     .create table GithubRaw (RunId:int, Version:string, OsType:string, Arch:string, DownloadCount:int, PublishDate: datetime, CountDate:datetime)
    - create mapping:   .create table ['GithubRawTest'] ingestion json mapping 'GithubRaw_mapping' '[{"column":"RunId", "Properties":{"Path":"$[\'RunId\']"}},{"column":"Version", "Properties":{"Path":"$[\'Version\']"}},{"column":"OsType", "Properties":{"Path":"$[\'OsType\']"}},{"column":"Arch", "Properties":{"Path":"$[\'Arch\']"}},{"column":"DownloadCount", "Properties":{"Path":"$[\'DownloadCount\']"}},{"column":"PublishDate", "Properties":{"Path":"$[\'PublishDate\']"}},{"column":"CountDate", "Properties":{"Path":"$[\'CountDate\']"}}]'

Github:
    - create table:     .create table Github (RunId:int, Version:string, DownloadCount:int, PublishDate: datetime, CountDate:datetime)
    - create mapping:  .create table ['Github'] ingestion json mapping 'Github_mapping' '[{"column":"RunId", "Properties":{"Path":"$[\'RunId\']"}},{"column":"Version", "Properties":{"Path":"$[\'Version\']"}},{"column":"DownloadCount", "Properties":{"Path":"$[\'DownloadCount\']"}},{"column":"PublishDate", "Properties":{"Path":"$[\'PublishDate\']"}},{"column":"CountDate", "Properties":{"Path":"$[\'CountDate\']"}}]'

HomeBrewRaw:
    - create table:     .create table HomebrewRaw (RunId:int, DataType:string, OsType: string, DownloadCount: int, CountDate:datetime)
    - create mapping:   .create table ['HomebrewRaw'] ingestion json mapping 'HomebrewRaw_mapping' '[{"column":"RunId", "Properties":{"Path":"$[\'RunId\']"}},{"column":"DataType", "Properties":{"Path":"$[\'DataType\']"}},{"column":"OsType", "Properties":{"Path":"$[\'OsType\']"}},{"column":"DownloadCount", "Properties":{"Path":"$[\'DownloadCount\']"}},{"column":"CountDate", "Properties":{"Path":"$[\'CountDate\']"}}]'

HomeBrew:
    - create table:     .create table Homebrew (RunId:int, OsType:string, DownloadCount:int, CountDate: datetime)
    - create mapping:   .create table ['Homebrew'] ingestion json mapping 'Homebrew_mapping' '[{"column":"RunId", "Properties":{"Path":"$[\'RunId\']"}},{"column":"OsType", "Properties":{"Path":"$[\'OsType\']"}},{"column":"DownloadCount", "Properties":{"Path":"$[\'DownloadCount\']"}},{"column":"CountDate", "Properties":{"Path":"$[\'CountDate\']"}}]'