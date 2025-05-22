package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/lzf-12/go-example-collections/api/graphql"
	"github.com/lzf-12/go-example-collections/api/grpc"
	"github.com/lzf-12/go-example-collections/api/rest"
)

const shutdowntimeout = 10 * time.Second

func main() {

	mode := flag.String("mode",
		"resthttp",
		"available mode: resthttp | restgin | restfiber | graphql | grpc")
	flag.Parse()
	serverMode := strings.ToLower(*mode)

	// channel for shutdown signal
	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, syscall.SIGINT, syscall.SIGTERM)

	serverErrs := make(chan error, 1) // buffer for at least 1 error
	var wg sync.WaitGroup

	switch serverMode {
	case "resthttp":
		startServer(&wg, serverErrs, func() error {
			return rest.ServeRestHttp()
		})
	case "restgin":
		startServer(&wg, serverErrs, func() error {
			return rest.ServeRestGin()
		})
	case "restfiber":
		startServer(&wg, serverErrs, func() error {
			return rest.ServeRestFiber()
		})
	case "graphql":
		startServer(&wg, serverErrs, func() error {
			return graphql.ServeGraphql()
		})
	case "grpc":
		startServer(&wg, serverErrs, func() error {
			return grpc.ServeGrpc()
		})
	default:
		log.Printf("%s. is invalid mode. valid mode are: rest, graphql, all", serverMode)
		os.Exit(1)
	}

	// handle graceful shutdown period
	shutdownDone := make(chan struct{})
	go func() {

		// add other component here to be terminated before shutdown
		// need to set wg.Add(1) per component

		wg.Wait()
		close(shutdownDone)
	}()

	// wait for shutdown signal
	<-shutdownSig
	log.Println("shutting down...")
	wg.Done()

	select {
	case <-shutdownDone:
		log.Println("clean shutdown")
	case <-time.After(shutdowntimeout):
		log.Println("shutdown process taking too long - forcing exit")
	case sig := <-shutdownSig: // second signal
		log.Printf("received %v - forcing exit", sig)
	}

	// check for catch errors
	select {
	case err := <-serverErrs:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %v", err)
			wg.Done()
		}
	default:
	}
}

// helper for consistently handling waitgroup counter on different mode
func startServer(wg *sync.WaitGroup, errChan chan<- error, serverFunc func() error) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		errChan <- serverFunc()
	}()
}
