package periodic_scanner

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockScannerQueue implements the ScannerQueue interface for testing
type MockScannerQueue struct {
	mock.Mock
}

func (m *MockScannerQueue) AddAllToQueue() error {
	return m.Called().Error(0)
}

func TestMain(m *testing.M) {
	test_utils.UnitTestRun(m)
}

func resetPeriodicScanner() {
	mainPeriodicScannerMutex.Lock()
	defer mainPeriodicScannerMutex.Unlock()

	if mainPeriodicScanner != nil {
		select {
		case <-mainPeriodicScanner.done:
			// Already closed
		default:
			close(mainPeriodicScanner.done)
		}
		if mainPeriodicScanner.ticker != nil {
			mainPeriodicScanner.ticker.Stop()
		}
		mainPeriodicScanner = nil
	}
}

func createTestSiteInfo(db *gorm.DB, interval int) error {
	siteInfo := models.SiteInfo{
		InitialSetup:         false,
		PeriodicScanInterval: interval,
		ConcurrentWorkers:    1,
	}
	return db.Create(&siteInfo).Error
}

func TestGetPeriodicScanInterval(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	t.Run("successful retrieval", func(t *testing.T) {
		assert.NoError(t, createTestSiteInfo(db, 300))

		duration, err := getPeriodicScanInterval(db)
		assert.NoError(t, err)
		assert.Equal(t, 300*time.Second, duration)
	})

	t.Run("database error - no site info", func(t *testing.T) {
		db.Exec("DELETE FROM site_info")

		duration, err := getPeriodicScanInterval(db)
		assert.Error(t, err)
		assert.Equal(t, time.Duration(0), duration)
	})
}

func TestInitializePeriodicScanner(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	t.Run("successful initialization with injection", func(t *testing.T) {
		defer resetPeriodicScanner()

		mockQueue := &MockScannerQueue{}
		assert.NoError(t, createTestSiteInfo(db, 300))
		assert.NoError(t, InitializePeriodicScannerWithQueue(db, mockQueue))

		// Verify initialization
		mainPeriodicScannerMutex.Lock()
		scanner := mainPeriodicScanner
		mainPeriodicScannerMutex.Unlock()

		assert.NotNil(t, scanner)
		assert.NotNil(t, scanner.scannerQueue)
		assert.Equal(t, mockQueue, scanner.scannerQueue)

		// Verify ticker is set up
		scanner.mutex.Lock()
		tickerExists := scanner.ticker != nil
		scanner.mutex.Unlock()
		assert.True(t, tickerExists)
	})

	t.Run("backward compatibility with original function", func(t *testing.T) {
		defer resetPeriodicScanner()

		assert.NoError(t, createTestSiteInfo(db, 300))
		assert.NoError(t, InitializePeriodicScanner(db))

		// Verify it uses RealScannerQueue
		mainPeriodicScannerMutex.Lock()
		defer mainPeriodicScannerMutex.Unlock()

		assert.NotNil(t, mainPeriodicScanner)
		assert.IsType(t, &RealScannerQueue{}, mainPeriodicScanner.scannerQueue)
	})

	t.Run("double initialization error", func(t *testing.T) {
		defer resetPeriodicScanner()

		mockQueue := &MockScannerQueue{}
		assert.NoError(t, createTestSiteInfo(db, 300))
		assert.NoError(t, InitializePeriodicScannerWithQueue(db, mockQueue))

		err := InitializePeriodicScannerWithQueue(db, mockQueue)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already been initialized")
	})
}

func TestScanIntervalRunnerWithMocking(t *testing.T) {
	t.Run("runner calls queue on ticker events", func(t *testing.T) {
		mockQueue := &MockScannerQueue{}

		// Use a channel to synchronize and count calls
		callChan := make(chan struct{}, 5) // Buffer for multiple calls
		mockQueue.On("AddAllToQueue").Return(nil).Run(func(args mock.Arguments) {
			select {
			case callChan <- struct{}{}:
			default:
				// Channel full, but that's okay
			}
		}).Maybe()

		ps := &periodicScanner{
			ticker:         time.NewTicker(50 * time.Millisecond),
			ticker_changed: make(chan bool, 1),
			done:           make(chan struct{}),
			mutex:          &sync.Mutex{},
			scannerQueue:   mockQueue,
		}

		// Start runner
		go ps.scanIntervalRunner()

		// Wait for at least one call with timeout
		select {
		case <-callChan:
			// Success - at least one call received
		case <-time.After(200 * time.Millisecond):
			t.Fatal("Expected at least one call to AddAllToQueue within timeout")
		}

		// Proper cleanup - stop ticker first, then close done
		ps.ticker.Stop()
		close(ps.done)

		// Give time for goroutine to finish
		time.Sleep(50 * time.Millisecond)

		// Verify the queue was called as expected
		mockQueue.AssertExpectations(t)
	})

	t.Run("runner handles queue errors gracefully", func(t *testing.T) {
		mockQueue := &MockScannerQueue{}
		// Mock queue to return an error
		mockQueue.On("AddAllToQueue").Return(errors.New("queue error")).Maybe()

		ps := &periodicScanner{
			ticker:         time.NewTicker(30 * time.Millisecond),
			ticker_changed: make(chan bool, 1),
			done:           make(chan struct{}),
			mutex:          &sync.Mutex{},
			scannerQueue:   mockQueue,
		}

		// Start runner and let it run briefly
		go ps.scanIntervalRunner()
		time.Sleep(80 * time.Millisecond)

		// Proper cleanup
		ps.ticker.Stop()
		close(ps.done)
		time.Sleep(50 * time.Millisecond)

		// Test passes if no panic occurred - errors should be logged gracefully
		mockQueue.AssertExpectations(t)
	})

	t.Run("runner responds to shutdown signal", func(t *testing.T) {
		mockQueue := &MockScannerQueue{}

		ps := &periodicScanner{
			ticker:         nil, // No ticker to avoid timing issues
			ticker_changed: make(chan bool, 1),
			done:           make(chan struct{}),
			mutex:          &sync.Mutex{},
			scannerQueue:   mockQueue,
		}

		runnerDone := make(chan bool)
		go func() {
			ps.scanIntervalRunner()
			close(runnerDone)
		}()

		close(ps.done)

		select {
		case <-runnerDone:
			// Success
		case <-time.After(1 * time.Second):
			t.Fatal("Runner did not exit within timeout")
		}

		mockQueue.AssertExpectations(t)
	})

	t.Run("runner responds to ticker changes", func(t *testing.T) {
		mockQueue := &MockScannerQueue{}

		ps := &periodicScanner{
			ticker:         nil, // Start without ticker
			ticker_changed: make(chan bool, 1),
			done:           make(chan struct{}),
			mutex:          &sync.Mutex{},
			scannerQueue:   mockQueue,
		}

		runnerDone := make(chan bool)
		go func() {
			ps.scanIntervalRunner()
			close(runnerDone)
		}()

		// Send ticker change signal
		select {
		case ps.ticker_changed <- true:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Could not send ticker change signal")
		}

		// Clean shutdown
		close(ps.done)

		// Wait for completion
		select {
		case <-runnerDone:
			// Success
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Runner did not exit after ticker change")
		}

		// Test passes if no deadlock occurs
		mockQueue.AssertExpectations(t)
	})
}

func TestRealScannerQueue(t *testing.T) {
	t.Run("real scanner queue interface compliance", func(t *testing.T) {
		queue := &RealScannerQueue{}

		// Just verify it implements the interface correctly
		var _ ScannerQueue = queue

		// Test that it doesn't panic when created
		assert.NotNil(t, queue)

		// We don't test the actual call since it requires external setup
	})
}
