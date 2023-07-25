package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"aztfy-download-counter/database"
	"aztfy-download-counter/datasource"
	"aztfy-download-counter/job"
	"aztfy-download-counter/job/githubutils"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

const DBName = "aztfy"
const HBContainer = "Homebrew"
const GHContainer = "Github"
const PMCContainer = "PMC"

func main() {
	ctx := context.TODO()
	cosmosdbEndPoint := os.Getenv("COSMOSDB_ENDPOINT")
	cosmosdbKey := os.Getenv("COSMOSDB_KEY")
	armClientId := os.Getenv("ARM_CLIENT_ID")
	armClientSecret := os.Getenv("ARM_CLIENT_SECRET")
	armTenantId := os.Getenv("ARM_TENANT_ID")
	pmcKustoEndpoint := os.Getenv("PMC_KUSTO_ENDPOINT")
	pmcStartDate := os.Getenv("PMC_START_DATE")

	standardDate := time.Now().UTC().Format(job.TimeFormat)

	dbClient, err := database.AuthDBClient(cosmosdbEndPoint, cosmosdbKey, DBName)
	if err != nil {
		log.Println(fmt.Errorf("init db client error: %+v", err))
	}

	logChan := make(chan string)
	go func(logChan chan string) {
		for message := range logChan {
			log.Print(message)
		}
	}(logChan)

	jobs := []job.Job{
		job.GithubWorker{
			Date: standardDate,
			ContainerInitFunc: func() (*azcosmos.ContainerClient, error) {
				return dbClient.NewContainer(GHContainer)
			},
			Logger: log.New(&logChanWriter{logChan: logChan}, "[GithubWorker]\t", 0),
		},
		job.HomebrewWorker{
			Date:   standardDate,
			Logger: log.New(&logChanWriter{logChan: logChan}, "[HomebrewWorker]\t", 0),
			ContainerInitFunc: func() (*azcosmos.ContainerClient, error) {
				return dbClient.NewContainer(HBContainer)
			},
			OsTypes: []database.OsType{
				database.OsTypeDarwin,
				database.OsTypeLinux,
			},
		},
	}

	if len(pmcStartDate) == 0 {
		pmcStartDate = standardDate
	}

	d, _ := time.Parse(job.TimeFormat, pmcStartDate)
	n, _ := time.Parse(job.TimeFormat, standardDate)
	cnt := n.Sub(d).Hours() / 24
	log.Println("PMC Start Date:", pmcStartDate, "Count:", int(cnt)+1)
	pmcJobs := []job.Job{}
	for i := 0; i <= int(cnt); i++ {
		pmcJobs = append(pmcJobs, job.PMCWorker{
			Date: d.Add(time.Hour * 24 * time.Duration(i)).Format(job.TimeFormat),
			ContainerInitFunc: func() (*azcosmos.ContainerClient, error) {
				return dbClient.NewContainer(PMCContainer)
			},
			KustoEndpoint:   pmcKustoEndpoint,
			Logger:          log.New(&logChanWriter{logChan: logChan}, "[PMCWorker]\t", 0),
			ArmClientId:     armClientId,
			ArmClientSecret: armClientSecret,
			ArmTenantId:     armTenantId,
		})
	}

	singlePmcRun := len(pmcJobs) > 1
	if !singlePmcRun {
		jobs = append(jobs, pmcJobs...)
	}

	var wg sync.WaitGroup
	for _, w := range jobs {
		wg.Add(1)

		go func(w job.Job) {
			defer wg.Done()
			w.Run(ctx)
		}(w)
	}

	if singlePmcRun {
		for _, job := range pmcJobs {
			job.Run(ctx)
		}
	}

	wg.Wait()
}

type logChanWriter struct {
	logChan chan<- string
}

func (w *logChanWriter) Write(p []byte) (int, error) {
	w.logChan <- string(p)
	return len(p), nil
}

func FetchGitHubVersionList(ctx context.Context) (map[string][]string, error) {
	releases, err := datasource.FetchGitHubDownloadCount(ctx)
	if err != nil {
		return nil, err
	}

	output := make(map[string][]string, 0)
	for _, r := range releases {
		for _, a := range r.Assets {
			if a.Name == nil || a.DownloadCount == nil {
				continue
			}

			version, _, arch, err := githubutils.ParseTagName(*a.Name, *a.ContentType)
			if err != nil {
				continue
			}

			if _, ok := output[version]; !ok {
				output[version] = make([]string, 0)
			}
			output[version] = append(output[version], arch)
		}
	}

	return output, nil
}
