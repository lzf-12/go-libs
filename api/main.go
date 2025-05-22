package main

import (
	"context"
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
	"github.com/lzf-12/go-example-collections/api/pubsub"
	"github.com/lzf-12/go-example-collections/api/rest"
)

const shutdowntimeout = 10 * time.Second

func main() {

	mode := flag.String("mode",
		"resthttp",
		"available mode: resthttp | restgin | restfiber | graphql | grpc | consumer-rabbitmq")
	flag.Parse()
	serverMode := strings.ToLower(*mode)

	// channel for shutdown signal
	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, syscall.SIGINT, syscall.SIGTERM)

	serverErrs := make(chan error, 1) // buffer for at least 1 error
	var wg sync.WaitGroup

	shutdownctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	switch serverMode {
	case "resthttp":
		go func() {
			err := rest.ServeRestHttp(shutdownctx)
			if err != nil {
				serverErrs <- err
			}
		}()
	case "restgin":
		go func() {
			err := rest.ServeRestGin(shutdownctx)
			if err != nil {
				serverErrs <- err
			}
		}()
	case "restfiber":
		go func() {
			err := rest.ServeRestFiber(shutdownctx)
			if err != nil {
				serverErrs <- err
			}
		}()
	case "graphql":
		go func() {
			err := graphql.ServeGraphql(shutdownctx)
			if err != nil {
				serverErrs <- err
			}
		}()
	case "grpc":
		go func() {
			err := grpc.ServeGrpc(shutdownctx)
			if err != nil {
				serverErrs <- err
			}
		}()
	case "consumer-rabbitmq":
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := pubsub.ServeRabbitMQConsumer(shutdownctx)
			if err != nil {
				serverErrs <- err
			}
		}()
	default:
		log.Printf("%s. is invalid mode. valid mode are: resthttp | restgin | restfiber | graphql | grpc | consumer-rabbitmq", serverMode)
		os.Exit(1)
	}

	// wait for shutdown signal
	<-shutdownSig
	log.Println("shutdown signal received...")

	cancel()
	log.Println("sending shutdown context to dependency...")

	// Wait for shutdown
	shutdownDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(shutdownDone)
	}()

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
