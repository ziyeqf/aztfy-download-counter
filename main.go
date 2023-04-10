package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"aztfy-download-counter/database"
	"aztfy-download-counter/job"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

const DBName = "aztfystatistics"
const HBContainer = "Homebrew"
const GHContainer = "Github"
const PMCContainer = "Pmc"

func main() {
	ctx := context.TODO()
	cosmosdbEndPoint := os.Getenv("COSMOSDB_ENDPOINT")
	cosmosdbKey := os.Getenv("COSMOSDB_KEY")

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

	kustoEndpoint := os.Getenv("PMC_KUSTO_ENDPOINT")
	pmcStartDate := os.Getenv("PMC_START_DATE")
	if len(pmcStartDate) == 0 {
		pmcStartDate = standardDate
	}

	d, _ := time.Parse(job.TimeFormat, pmcStartDate)
	n, _ := time.Parse(job.TimeFormat, standardDate)
	cnt := n.Sub(d).Hours() / 24
	log.Println("PMC Start Date:", pmcStartDate, "Count:", int(cnt))
	for i := 0; i <= int(cnt); i++ {
		jobs = append(jobs, job.PMCWorker{
			Date: d.Add(time.Hour * 24 * time.Duration(i)).Format(job.TimeFormat),
			ContainerInitFunc: func() (*azcosmos.ContainerClient, error) {
				return dbClient.NewContainer(PMCContainer)
			},
			KustoEndpoint: kustoEndpoint,
			Logger:        log.New(&logChanWriter{logChan: logChan}, "[PMCWorker]\t", 0),
		})
	}

	var wg sync.WaitGroup
	for _, w := range jobs {
		wg.Add(1)

		go func(w job.Job) {
			defer wg.Done()
			w.Run(ctx)
		}(w)
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
