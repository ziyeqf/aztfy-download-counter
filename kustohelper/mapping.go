package kustohelper

const GithubIngestMapping string = `[{"column":"Version", "Properties":{"Path":"$[\'Version\']"}},{"column":"OsType", "Properties":{"Path":"$[\'OsType\']"}},{"column":"Arch", "Properties":{"Path":"$[\'Arch\']"}},{"column":"DownloadCount", "Properties":{"Path":"$[\'DownloadCount\']"}},{"column":"PublishDate", "Properties":{"Path":"$[\'PublishDate\']"}},{"column":"CountDate", "Properties":{"Path":"$[\'CountDate\']"}}]`
const HomeBrewRawIngestMapping string = `[{"column":"DataType", "Properties":{"Path":"$[\'DataType\']"}},{"column":"OsType", "Properties":{"Path":"$[\'OsType\']"}},{"column":"DownloadCount", "Properties":{"Path":"$[\'DownloadCount\']"}},{"column":"CountDate", "Properties":{"Path":"$[\'CountDate\']"}}]`
const HomeBrewIngestMapping string = `[{"column":"OsType", "Properties":{"Path":"$[\'OsType\']"}},{"column":"DownloadCount", "Properties":{"Path":"$[\'DownloadCount\']"}},{"column":"CountDate", "Properties":{"Path":"$[\'CountDate\']"}}]`
