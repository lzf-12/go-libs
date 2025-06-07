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

	"github.com/lzf-12/go-example-collections/internal/api/graphql"
	"github.com/lzf-12/go-example-collections/internal/api/grpc"
	"github.com/lzf-12/go-example-collections/internal/api/rest"
	"github.com/lzf-12/go-example-collections/internal/consumer"
)

const shutdowntimeout = 10 * time.Second

func main() {

	mode := flag.String("mode",
		"resthttp",
		"available mode: resthttp | restgin | restfiber | graphql | grpc | consumer-rabbitmq | consumer-kafka")
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
			if err := rest.ServeRestHttp(shutdownctx); err != nil {
				serverErrs <- err
				shutdownSig <- os.Interrupt
			}
		}()
	case "restgin":
		go func() {
			if err := rest.ServeRestGin(shutdownctx); err != nil {
				serverErrs <- err
				shutdownSig <- os.Interrupt
			}
		}()
	case "restfiber":
		go func() {
			if err := rest.ServeRestFiber(shutdownctx); err != nil {
				serverErrs <- err
				shutdownSig <- os.Interrupt
			}
		}()
	case "graphql":
		go func() {
			err := graphql.ServeGraphql(shutdownctx)
			if err != nil {
				serverErrs <- err
				shutdownSig <- os.Interrupt
			}
		}()
	case "grpc":
		go func() {
			err := grpc.ServeGrpc(shutdownctx)
			if err != nil {
				serverErrs <- err
				shutdownSig <- os.Interrupt
			}
		}()
	case "consumer-rabbitmq":
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := consumer.ServeRabbitMQConsumer(shutdownctx); err != nil {
				serverErrs <- err
				shutdownSig <- os.Interrupt
			}
		}()
	case "consumer-kafka":
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := consumer.ServeKafkaConsumer(shutdownctx); err != nil {
				serverErrs <- err
				shutdownSig <- os.Interrupt
			}
		}()
	default:
		log.Printf("%s. is invalid mode. valid mode are: resthttp | restgin | restfiber | graphql | grpc | consumer-rabbitmq | consumer-kafka", serverMode)
		os.Exit(1)
	}

	// wait for shutdown signal
	<-shutdownSig
	log.Println("shutdown signal received...")

	cancel()
	log.Println("sending shutdown context to dependency...")

	// wait for dependency shutdown
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
		}
	default:
	}
}
