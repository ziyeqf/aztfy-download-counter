package main

import (
	"context"
	"fmt"
	"log"
	"os"
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
		logError(err)
	}

	ghWorker := worker.GithubWorker{
		Container: ghContainer,
	}
	err = ghWorker.Run(ctx, standardDate)
	if err != nil {
		logError(fmt.Errorf("[Github] %+v", err))
	}
	log.Println("[Github] done")

	hbContainer, err := dbClient.NewContainer(HBContainer)
	if err != nil {
		logError(err)
	}

	hbWorker := worker.HomebrewWorker{
		Container: hbContainer,
		OsTypes: []database.OsType{
			database.OsTypeDarwin,
			database.OsTypeLinux,
		},
	}
	err = hbWorker.Run(ctx, standardDate)

	if err != nil {
		logError(fmt.Errorf("[Homebrew] %+v", err))
	}
	log.Println("[Homebrew] done")
}

func logError(err error) {
	log.Println(err)
	os.Exit(1)
}
