package profiler

import (
	"log"
	"runtime"
	"time"
)

// TimedMemStats represents a MemStats reading with the time it was taken
type TimedMemStats struct {
	// the unix time (seconds) the stats were polled
	TimeEpochMs int64

	// memory statistics
	MemStats runtime.MemStats
}

// HeapMemStat represents a snapshot of the heap memory at a given time.
type HeapMemStat struct {
	TimeMsAgo      int64 // ms ago (negative)
	SysKb          uint64
	HeapSysKb      uint64
	HeapAllocKb    uint64
	HeapIdleKb     uint64
	HeapReleasedKb uint64
}

// Convert the input slice of TimedMemStats to a slice of HeapMemStats
func timedMemStatsToHeapMemStats(timedMemStats []TimedMemStats) []HeapMemStat {
	result := make([]HeapMemStat, len(timedMemStats))
	nowMs := time.Now().UnixNano() / 1000000
	for index := range timedMemStats {
		result[index] = HeapMemStat{
			TimeMsAgo:      timedMemStats[index].TimeEpochMs - nowMs,
			SysKb:          timedMemStats[index].MemStats.Sys / uint64(1000),
			HeapSysKb:      timedMemStats[index].MemStats.HeapSys / uint64(1000),
			HeapAllocKb:    timedMemStats[index].MemStats.HeapAlloc / uint64(1000),
			HeapIdleKb:     timedMemStats[index].MemStats.HeapIdle / uint64(1000),
			HeapReleasedKb: timedMemStats[index].MemStats.HeapReleased / uint64(1000),
		}
	}
	return result
}

// TrackMemoryStatistics keeps track of runtime.MemStats.
// Sample memory statistics every [sampleIntervalMs] milliseconds, keeping the last [bufferSize]
// samples. Returns a request/response channel to get recent memory statistics and a quit channel
// to stop polling.
// Parameters:
// 		bufferSize: 				number of samples to keep
//		sampleIntervalMs: 			number of milliseconds to wait between polling
//		memStatsRequestChannel:		request/reply channel to report back the most recent N samples
// 		quitChannel: 				once closed, all polling stops
func TrackMemoryStatistics(bufferSize int, sampleIntervalMs int64, memStatsRequestChannel <-chan chan []TimedMemStats, quitChannel chan bool) {
	// channel to receive profiling data - with a little bit of a buffer for when we're responding to a request
	memStatsReceiveChannel := make(chan TimedMemStats, 20)

	log.Print("Starting memory profiling goroutines")

	// start polling
	go func() {
		for {
			select {
			case <-time.After(time.Duration(sampleIntervalMs) * time.Millisecond):

				// take a snapshot, send it into the channel
				var stats TimedMemStats
				runtime.ReadMemStats(&stats.MemStats)
				stats.TimeEpochMs = time.Now().UnixNano() / 1000000

				memStatsReceiveChannel <- stats

			case <-quitChannel:
				log.Print("Stopping memory profiling goroutine (1 of 2)")
				return
			}
		}
	}()

	// listen for new data coming in, and requests for all the data we currently have
	go func() {

		var memStats TimedMemStats
		var responseChan chan []TimedMemStats

		// Ring buffer for the most recent samples.
		// This is kept simple by only adding to it, never removing.
		sampleData := make([]TimedMemStats, bufferSize)

		// number of samples so far
		size := 0

		// tail index of the sample data - new data written to tail
		tail := -1

		for {
			select {
			case memStats = <-memStatsReceiveChannel:
				// received new sampling data - increment tail, wrapping around if necessary
				tail++
				if tail >= bufferSize {
					tail -= bufferSize
				}
				if size < bufferSize {
					size++
				}
				sampleData[tail] = memStats

			case responseChan = <-memStatsRequestChannel:
				// received a request for tracking data
				response := make([]TimedMemStats, size)

				writeIndex := 0

				// if we're currently wrapped around, read from tail+1 to the end of sampleData
				if size == bufferSize && tail < (bufferSize-1) {
					// copy everything after the tail to the end of the array
					for i := tail + 1; i < bufferSize; i++ {
						response[writeIndex] = sampleData[i]
						writeIndex++
					}
				}
				// add on everything from 0->tail
				for i := 0; i <= tail; i++ {
					response[writeIndex] = sampleData[i]
					writeIndex++
				}

				// send the response
				responseChan <- response
			case <-quitChannel:
				// We're all done
				log.Print("Stopping memory profiling goroutine (2 of 2)")
				return
			}
		}
	}()
}
