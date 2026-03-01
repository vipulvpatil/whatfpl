package fpl

import (
	"fmt"
	"log"
	"math/rand"
	"sync/atomic"
	"time"
)

const (
	refreshInterval = 5 * time.Minute
	refreshJitter   = 30 * time.Second
)

type DataManager struct {
	store atomic.Pointer[Store]
}

func NewDataManager() (*DataManager, error) {
	dm := &DataManager{}
	if err := dm.refresh(); err != nil {
		return nil, fmt.Errorf("initial fetch failed: %w", err)
	}
	go dm.loop()
	return dm, nil
}

func (dm *DataManager) Store() *Store {
	return dm.store.Load()
}

func (dm *DataManager) refresh() error {
	data, err := FetchBootstrap()
	if err != nil {
		return err
	}
	s, err := NewStore(data)
	if err != nil {
		return err
	}
	dm.store.Store(s)
	return nil
}

func (dm *DataManager) loop() {
	for {
		jitter := time.Duration(rand.Int63n(int64(refreshJitter)))
		time.Sleep(refreshInterval + jitter)
		if err := dm.refresh(); err != nil {
			log.Printf("fpl: refresh failed: %v", err)
		} else {
			log.Printf("fpl: store refreshed (gameweek %d, %d players)", dm.Store().CurrentGameweek, len(dm.Store().Players))
		}
	}
}
