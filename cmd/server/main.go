package main

import (
	"log"
	"net/http"

	"github.com/vipulvpatil/whatfpl/internal/api"
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

	handler := api.NewHandler(dm)
	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
