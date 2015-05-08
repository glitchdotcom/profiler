package profiler

import (
	"testing"
	"time"
)

func TestProfilingWhenNotFillingBuffer(t *testing.T) {
	t.Parallel()

	memStatsRequestChannel := make(chan chan []TimedMemStats)
	memStatsQuitChannel := make(chan bool)

	// 1 second of data, 10/sec
	TrackMemoryStatistics(10, 1000/10, memStatsRequestChannel, memStatsQuitChannel)
	defer close(memStatsQuitChannel)

	// wait 0.55 seconds
	time.Sleep(550 * time.Millisecond)

	// get the data
	responseChan := make(chan []TimedMemStats, 1)
	memStatsRequestChannel <- responseChan
	response := <-responseChan

	// make sure there's only 5 samplings, even though we gave it long enough for 14-15
	val := len(response)
	if val != 5 {
		t.Error("len(response) - Expected: 5, got", val)
	}

	// make sure they're in order - this tests that the ring buffer correctly flattened the results
	var prevTime int64
	for i := 0; i < 5; i++ {
		if prevTime >= response[i].TimeEpochMs {
			t.Error("results out of order")
		}
		prevTime = response[i].TimeEpochMs
	}
}

func TestProfilingWhenFillsBuffer(t *testing.T) {
	t.Parallel()

	memStatsRequestChannel := make(chan chan []TimedMemStats)
	memStatsQuitChannel := make(chan bool)

	// 1 second of data, 10 per sec
	TrackMemoryStatistics(10, 1000/10, memStatsRequestChannel, memStatsQuitChannel)
	defer close(memStatsQuitChannel)

	// wait 1.5 seconds
	time.Sleep(1500 * time.Millisecond)

	// get the data
	responseChan := make(chan []TimedMemStats, 1)
	memStatsRequestChannel <- responseChan
	response := <-responseChan

	// make sure there's only 10 samplings, even though we gave it long enough for 14-15
	val := len(response)
	if val != 10 {
		t.Error("len(response) - Expected: 10, got", val)
	}

	// make sure they're in order - this tests that the ring buffer correctly flattened the results
	var prevTime int64
	for i := 0; i < 10; i++ {
		if prevTime >= response[i].TimeEpochMs {
			t.Error("results out of order")
		}
		prevTime = response[i].TimeEpochMs
	}
}
