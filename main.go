package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func superCoolHandler(w http.ResponseWriter, r *http.Request) {

	// Start work. This can take anywhere from 1 to 5 seconds.
	results := make(chan string)
	go func() {
		delay := time.Duration(rand.Intn(4000)+1000) * time.Millisecond
		time.Sleep(delay)
		results <- fmt.Sprintf("Super cool answer after %v", delay)
	}()

	// Wait for the work to be done or the request to cancel.
	select {
	case result := <-results:
		io.WriteString(w, result)

	case <-r.Context().Done():
		log.Printf("Could not get super cool answer: %v", r.Context().Err())
	}
}

func main() {

	//////////////////////////////////////////////////////////////////////////////
	// Start a diagnostic server listening on an alternate port.
	// Not concerned with shutting this down gracefully.
	go func() {
		log.Println("Diagnostic server listening on localhost:6060")
		err := http.ListenAndServe("localhost:6060", http.DefaultServeMux)
		log.Printf("Diagnostic server error: %v", err)
	}()

	//////////////////////////////////////////////////////////////////////////////
	// Define application routes
	router := http.NewServeMux()
	router.HandleFunc("/", superCoolHandler)

	// Wrap router in middleware. This one from net/http puts a global maximum
	// time for each handler and will cancel requests' contexts.
	h := http.TimeoutHandler(router, 3*time.Second, "TIMEOUT")

	srv := http.Server{
		Addr:         "localhost:8080",
		Handler:      h,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	serveGracefully(&srv, 5*time.Second)
}

// serveGracefully starts serving srv on its defined Addr and blocks until it receives
func serveGracefully(srv *http.Server, timeout time.Duration) {

	// Start http server in another goroutine so it doesn't block.
	serverError := make(chan error, 1)
	go func() {
		log.Println("Application server listening on", srv.Addr)
		serverError <- srv.ListenAndServe()
	}()

	//////////////////////////////////////////////////////////////////////////////
	// Register this channel to be notified of shutdown signals.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	//////////////////////////////////////////////////////////////////////////////
	// Block main until we receive on one of these two channels.
	select {
	case err := <-serverError:
		log.Printf("Error starting server: %v", err)

	case sig := <-shutdown:
		log.Printf("Received %v signal, shutting down", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Could not shut down gracefully: %v", err)
			log.Println("Forcing shutdown...")
			log.Println(srv.Close())
		}
	}
}
