// Example service with the profiler listening on its only HTTP IP:port
package main

import (
	"fmt"
	"github.com/wblakecaldwell/profiler"
	"log"
	"net/http"
)

func main() {
	// add our "Hello, World!" endpoint to the default ServeMux
	http.HandleFunc("/", helloHandler)

	// add the profiler endpoints to the default ServeMux
	profiler.AddMemoryProfilingHandlers()

	// start the service
	log.Println("Starting service on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// simple handler that just says Hello
func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}
