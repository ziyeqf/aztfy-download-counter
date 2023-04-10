package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"aztfy-download-counter/database"
	"aztfy-download-counter/worker"
)

const DBName = "aztfystatistics"
const HBContainer = "Homebrew"
const GHContainer = "Github"

func main() {
	ctx := context.TODO()
	cosmosdbEndPoint := os.Getenv("COSMOSDB_ENDPOINT")
	cosmosdbKey := os.Getenv("COSMOSDB_KEY")

	standardDate := time.Now().UTC().Format(worker.TimeFormat)

	dbClient, err := database.AuthDBClient(cosmosdbEndPoint, cosmosdbKey, DBName)

	ghContainer, err := dbClient.NewContainer(GHContainer)
	if err != nil {
		log.Println(err)
	}

	hbContainer, err := dbClient.NewContainer(HBContainer)
	if err != nil {
		log.Println(err)
	}

	logChan := make(chan string)
	go func(logChan chan string) {
		for message := range logChan {
			log.Print(message)
		}
	}(logChan)

	workers := []worker.Worker{
		worker.GithubWorker{
			Logger:    log.New(&logChanWriter{logChan: logChan}, "[GithubWorker]\t", 0),
			Container: ghContainer,
		},
		worker.HomebrewWorker{
			Logger:    log.New(&logChanWriter{logChan: logChan}, "[HomebrewWorker]\t", 0),
			Container: hbContainer,
			OsTypes: []database.OsType{
				database.OsTypeDarwin,
				database.OsTypeLinux,
			},
		},
	}

	var wg sync.WaitGroup
	for _, w := range workers {
		wg.Add(1)

		go func(w worker.Worker) {
			defer wg.Done()
			w.Run(ctx, standardDate)
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
