// Example service with the profiler listening on a separate port from other HTTP endpoints
package main

import (
	"fmt"
	"github.com/wblakecaldwell/profiler"
	"log"
	"net/http"
)

func main() {
	// start the profiler on its own port
	setupProfiler(":6060")

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
	mux.HandleFunc("/profiler/info.html", profiler.MemStatsHTMLHandler)
	mux.HandleFunc("/profiler/info", profiler.ProfilingInfoJSONHandler)
	mux.HandleFunc("/profiler/start", profiler.StartProfilingHandler)
	mux.HandleFunc("/profiler/stop", profiler.StopProfilingHandler)

	log.Printf("Starting profiler on %s\n", listen)
	go func() {
		log.Fatal(http.ListenAndServe(listen, mux))
	}()
}
