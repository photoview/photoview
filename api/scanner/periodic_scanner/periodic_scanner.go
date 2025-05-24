package periodic_scanner

import (
	"fmt"
	"sync"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner/scanner_queue"
	"gorm.io/gorm"
)

type periodicScanner struct {
	ticker         *time.Ticker
	ticker_changed chan bool
	done           chan struct{}
	mutex          *sync.Mutex
	db             *gorm.DB
}

var mainPeriodicScannerMutex sync.Mutex
var mainPeriodicScanner *periodicScanner = nil

func getPeriodicScanInterval(db *gorm.DB) (time.Duration, error) {
	var siteInfo models.SiteInfo
	if err := db.First(&siteInfo).Error; err != nil {
		return 0, err
	}

	return time.Duration(siteInfo.PeriodicScanInterval) * time.Second, nil
}

func InitializePeriodicScanner(db *gorm.DB) error {
	mainPeriodicScannerMutex.Lock()
	defer mainPeriodicScannerMutex.Unlock()

	if mainPeriodicScanner != nil {
		return fmt.Errorf("periodic scanner has already been initialized")
	}

	scanInterval, err := getPeriodicScanInterval(db)
	if err != nil {
		return err
	}

	mainPeriodicScanner = &periodicScanner{
		db:             db,
		ticker_changed: make(chan bool),
		done:           make(chan struct{}),
		mutex:          &sync.Mutex{},
	}

	go mainPeriodicScanner.scanIntervalRunner()

	ChangePeriodicScanInterval(scanInterval)
	return nil
}

func ChangePeriodicScanInterval(duration time.Duration) {
	var newTicker *time.Ticker = nil
	if duration > 0 {
		newTicker = time.NewTicker(duration)
		log.Info("Periodic scan interval changed: %s", duration.String())
	} else {
		log.Info("Periodic scan interval changed: disabled")
	}

	mainPeriodicScannerMutex.Lock()
	scanner := mainPeriodicScanner
	mainPeriodicScannerMutex.Unlock()

	if scanner != nil {
		scanner.mutex.Lock()
		defer scanner.mutex.Unlock()

		if scanner.ticker != nil {
			scanner.ticker.Stop()
		}

		scanner.ticker = newTicker
		scanner.ticker_changed <- true
	}
}

// ShutdownPeriodicScanner gracefully shuts down the periodic scanner
func ShutdownPeriodicScanner() {
	mainPeriodicScannerMutex.Lock()
	defer mainPeriodicScannerMutex.Unlock()

	if mainPeriodicScanner != nil {
		log.Info("Shutting down periodic scanner")

		// Signal the runner goroutine to stop
		close(mainPeriodicScanner.done)

		// Stop the ticker if it exists
		mainPeriodicScanner.mutex.Lock()
		if mainPeriodicScanner.ticker != nil {
			mainPeriodicScanner.ticker.Stop()
			mainPeriodicScanner.ticker = nil
		}
		mainPeriodicScanner.mutex.Unlock()

		// Reset the global scanner
		mainPeriodicScanner = nil
	}
}

func (ps *periodicScanner) scanIntervalRunner() {
	for {
		log.Info("Scan interval runner: Waiting for signal")

		ps.mutex.Lock()
		ticker := ps.ticker
		ps.mutex.Unlock()

		if ticker != nil {
			select {
			case <-ps.done:
				log.Info("Scan interval runner: Shutting down")
				return
			case <-ps.ticker_changed:
				log.Info("Scan interval runner: New ticker detected")
			case <-ticker.C:
				log.Info("Scan interval runner: Starting periodic scan")
				if err := scanner_queue.AddAllToQueue(); err != nil {
					log.Error("Scan interval runner: Failed to add all users to queue: %v", err)
				}
			}
		} else {
			select {
			case <-ps.done:
				log.Info("Scan interval runner: Shutting down")
				return
			case <-ps.ticker_changed:
				log.Info("Scan interval runner: New ticker detected")
			}
		}
	}
}
