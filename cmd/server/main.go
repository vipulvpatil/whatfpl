package main

import (
	"log"

	"github.com/vipulvpatil/whatfpl/internal/fpl"
)

func main() {
	log.Println("starting whatfpl...")

	dm, err := fpl.NewDataManager()
	if err != nil {
		log.Fatalf("failed to initialize data manager: %v", err)
	}

	store := dm.Store()
	log.Printf("loaded gameweek %d with %d players and %d teams",
		store.CurrentGameweek, len(store.Players), len(store.Teams))

	// Block forever — API server goes here later.
	select {}
}
