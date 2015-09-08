Example: Profiler on a Separate Port
====================================

If you're adding the profiler to a service that serves HTTP endpoints, you might
want the profiler listening on its own IP/port. This makes it easier to keep
your profiling information private, and set up firewall rules to not serve this
IP/port publicly.

The trick here is to not use the DefaultServeMux. This is all you need to add to 
your project to set up the profiler to listen on its own port:

```go
// Set up the profiler to listen on the input port:IP string
func setupProfiler(listen string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/profiler/info.html", profiler.MemStatsHTMLHandler)
	mux.HandleFunc("/profiler/info", profiler.ProfilingInfoJSONHandler)
	mux.HandleFunc("/profiler/start", profiler.StartProfilingHandler)
	mux.HandleFunc("/profiler/stop", profiler.StopProfilingHandler)
	fmt.Printf("Starting profiler on %s\n", listen)
	go http.ListenAndServe(listen, mux)
}
```


Run the Example
---------------

Fetch, build, and run the example service:

```shell
go get github.com/wblakecaldwell/profiler
go build github.com/wblakecaldwell/profiler/examples/separate_port
./separate_port
```

Verify the 'Hello, World!' endpoint at http://localhost:8080

Verify the profiler is running on its own port at http://localhost:6060/profiler/info.html

![Screenshot](screenshot.png)
