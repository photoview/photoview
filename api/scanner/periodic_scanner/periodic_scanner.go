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

type ScannerQueue interface {
	AddAllToQueue() error
}

type RealScannerQueue struct{}

func (r *RealScannerQueue) AddAllToQueue() error {
	return scanner_queue.AddAllToQueue()
}

type periodicScanner struct {
	ticker         *time.Ticker
	tickerLocker   sync.Mutex
	ticker_changed chan bool
	done           chan struct{}
	db             *gorm.DB
	scannerQueue   ScannerQueue
}

var mainPeriodicScanner *periodicScanner = nil

func getPeriodicScanInterval(db *gorm.DB) (time.Duration, error) {
	var siteInfo models.SiteInfo
	if err := db.First(&siteInfo).Error; err != nil {
		return 0, err
	}

	return time.Duration(siteInfo.PeriodicScanInterval) * time.Second, nil
}

func InitializePeriodicScannerWithQueue(db *gorm.DB, queue ScannerQueue) error {

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
		tickerLocker:   sync.Mutex{},
		scannerQueue:   queue,
	}

	go mainPeriodicScanner.scanIntervalRunner()

	var newTicker *time.Ticker = nil
	if scanInterval > 0 {
		newTicker = time.NewTicker(scanInterval)
		log.Info(nil, "Periodic scan interval changed: "+scanInterval.String())
	} else {
		log.Info(nil, "Periodic scan interval changed: disabled")
	}

	mainPeriodicScanner.ticker = newTicker

	select {
	case mainPeriodicScanner.ticker_changed <- true:
	default:
		// Channel might be full, but that's okay
	}

	return nil
}

func InitializePeriodicScanner(db *gorm.DB) error {
	return InitializePeriodicScannerWithQueue(db, &RealScannerQueue{})
}

func ChangePeriodicScanInterval(duration time.Duration) {
	var newTicker *time.Ticker = nil
	if duration > 0 {
		newTicker = time.NewTicker(duration)
		log.Info(nil, "Periodic scan interval changed: "+duration.String())
	} else {
		log.Info(nil, "Periodic scan interval changed: disabled")
	}

	scanner := mainPeriodicScanner
	if scanner != nil {
		scanner.tickerLocker.Lock()
		defer scanner.tickerLocker.Unlock()

		if scanner.ticker != nil {
			scanner.ticker.Stop()
		}

		scanner.ticker = newTicker
		select {
		case scanner.ticker_changed <- true:
		default:
			// Channel might be full, but that's okay
		}
	}
}

// ShutdownPeriodicScanner gracefully shuts down the periodic scanner
func ShutdownPeriodicScanner() {

	if mainPeriodicScanner != nil {
		log.Info(nil, "Shutting down periodic scanner")

		// Signal the runner goroutine to stop
		close(mainPeriodicScanner.done)

		// Stop the ticker if it exists
		mainPeriodicScanner.tickerLocker.Lock()
		if mainPeriodicScanner.ticker != nil {
			mainPeriodicScanner.ticker.Stop()
			mainPeriodicScanner.ticker = nil
		}
		mainPeriodicScanner.tickerLocker.Unlock()

		// Reset the global scanner
		mainPeriodicScanner = nil
	}
}

func (ps *periodicScanner) scanIntervalRunner() {
	for {
		log.Info(nil, "Scan interval runner: Waiting for signal")

		ps.tickerLocker.Lock()
		ticker := ps.ticker
		ps.tickerLocker.Unlock()

		if ticker != nil {
			select {
			case <-ps.done:
				log.Info(nil, "Scan interval runner: Shutting down")
				return
			case <-ps.ticker_changed:
				log.Info(nil, "Scan interval runner: New ticker detected")
			case <-ticker.C:
				log.Info(nil, "Scan interval runner: Starting periodic scan")
				if err := ps.scannerQueue.AddAllToQueue(); err != nil {
					log.Error(nil, "Scan interval runner: Failed to add all users to queue", "error", err)
				}
			}
		} else {
			select {
			case <-ps.done:
				log.Info(nil, "Scan interval runner: Shutting down")
				return
			case <-ps.ticker_changed:
				log.Info(nil, "Scan interval runner: New ticker detected")
			}
		}
	}
}
