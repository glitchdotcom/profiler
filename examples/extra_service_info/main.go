// Example service with the profiler listening on a separate port,
// and sharing extra diagnostic info as key/value pairs
package main

import (
	"fmt"
	"github.com/wblakecaldwell/profiler"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

// variables for extra service info
var (
	// startTime is the time that the service started
	startTime time.Time

	// infoHtmlHitCount keeps track of how many times /profiler/info.html is hit
	infoHTMLHitCount uint64

	// infoHitCount keeps track of how many times the /profiler/info is hit
	infoHitCount uint64
)

func init() {
	startTime = time.Now().Round(time.Second)
}

func main() {
	// start the profiler on its own port
	setupProfiler(":6060")

	// provide the profiler with a function where it can request key/value pairs as extra service info
	profiler.RegisterExtraServiceInfoRetriever(extraServiceInfo)

	// Serve your public HTTP endpoints on port 8080
	http.HandleFunc("/", helloHandler)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// simple handler that just says Hello
func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

// Set up the profiler to listen on the input port:IP string
func setupProfiler(listen string) {
	mux := http.NewServeMux()

	// wrap /profiler/info.html for hit tracking
	mux.HandleFunc("/profiler/info.html", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&infoHTMLHitCount, 1)
		profiler.MemStatsHTMLHandler(w, r)
	})

	// wrap /profiler/info for hit tracking
	mux.HandleFunc("/profiler/info", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&infoHitCount, 1)
		profiler.ProfilingInfoJSONHandler(w, r)
	})

	mux.HandleFunc("/profiler/start", profiler.StartProfilingHandler)
	mux.HandleFunc("/profiler/stop", profiler.StopProfilingHandler)
	log.Printf("Starting profiler on %s\n", listen)
	go func() {
		log.Fatal(http.ListenAndServe(listen, mux))
	}()
}

// extraServiceInfo implements the profiler.ExtraServiceInfoRetriever interface,
// returning a map of key/value pairs of diagnostic information.
func extraServiceInfo() map[string]interface{} {
	extraInfo := make(map[string]interface{})
	extraInfo["uptime"] = time.Now().Round(time.Second).Sub(startTime).String()
	extraInfo["hit count: /profiler/info.html"] = atomic.LoadUint64(&infoHTMLHitCount)
	extraInfo["hit count: /profiler/info"] = atomic.LoadUint64(&infoHitCount)
	return extraInfo
}
