package periodic_scanner

import (
	"log"
	"sync"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_queue"
	"gorm.io/gorm"
)

type periodicScanner struct {
	ticker         *time.Ticker
	ticker_changed chan bool
	mutex          *sync.Mutex
	db             *gorm.DB
}

var mainPeriodicScanner *periodicScanner = nil

func getPeriodicScanInterval(db *gorm.DB) (time.Duration, error) {

	var siteInfo models.SiteInfo
	if err := db.First(&siteInfo).Error; err != nil {
		return 0, err
	}

	return time.Duration(siteInfo.PeriodicScanInterval) * time.Second, nil
}

func InitializePeriodicScanner(db *gorm.DB) error {
	if mainPeriodicScanner != nil {
		panic("periodic scanner has already been initialized")
	}

	scanInterval, err := getPeriodicScanInterval(db)
	if err != nil {
		return err
	}

	mainPeriodicScanner = &periodicScanner{
		db:             db,
		ticker_changed: make(chan bool),
		mutex:          &sync.Mutex{},
	}

	go scanIntervalRunner()

	ChangePeriodicScanInterval(scanInterval)
	return nil
}

func ChangePeriodicScanInterval(duration time.Duration) {
	var newTicker *time.Ticker = nil
	if duration > 0 {
		newTicker = time.NewTicker(duration)
		log.Printf("Periodic scan interval changed: %s", duration.String())
	} else {
		log.Print("Periodic scan interval changed: disabled")
	}

	{
		mainPeriodicScanner.mutex.Lock()
		defer mainPeriodicScanner.mutex.Unlock()

		if mainPeriodicScanner.ticker != nil {
			mainPeriodicScanner.ticker.Stop()
		}

		mainPeriodicScanner.ticker = newTicker
		mainPeriodicScanner.ticker_changed <- true
	}
}

func scanIntervalRunner() {
	for {
		log.Print("Scan interval runner: Waiting for signal")
		if mainPeriodicScanner.ticker != nil {
			select {
			case <-mainPeriodicScanner.ticker_changed:
				log.Print("Scan interval runner: New ticker detected")
			case <-mainPeriodicScanner.ticker.C:
				log.Print("Scan interval runner: Starting periodic scan")
				scanner_queue.AddAllToQueue()
			}
		} else {
			<-mainPeriodicScanner.ticker_changed
			log.Print("Scan interval runner: New ticker detected")
		}
	}
}
